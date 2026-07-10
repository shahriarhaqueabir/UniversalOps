package common

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// MetricWrite represents a single metric enqueued for asynchronous batch write.
type MetricWrite struct {
	Name  string
	Unit  string
	Value float64
	Time  time.Time
}

// Storage manages the persistent SQLite database.
type Storage struct {
	db         *sql.DB
	insertStmt *sql.Stmt
	metricsCh  chan MetricWrite
	closeCh    chan struct{}
	writerWg   sync.WaitGroup
	pruneWg    sync.WaitGroup
	flushCh    chan chan struct{}
}

var (
	// DefaultDBName is the name of the SQLite database file.
	DefaultDBName = "hawkward.db"
	globalStorage *Storage
)

// InitStorage initializes the global SQLite storage.
func InitStorage(path string) error {
	if path == "" {
		path = DefaultDBName
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("create storage dir: %w", err)
		}
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return fmt.Errorf("open sqlite: %w", err)
	}

	// Performance settings: single connection since SQLite is file-level locked
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	// WAL mode and performance pragmas
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=-8000",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return fmt.Errorf("pragma %s: %w", p, err)
		}
	}

	s := &Storage{
		db:        db,
		metricsCh: make(chan MetricWrite, 256),
		closeCh:   make(chan struct{}),
		flushCh:   make(chan chan struct{}),
	}

	// Migrate first to ensure tables exist, then prepare statements
	if err := s.migrate(); err != nil {
		db.Close()
		return fmt.Errorf("storage migration: %w", err)
	}

	// Prepare the insert statement for batched metric writes
	s.insertStmt, err = db.Prepare(`INSERT INTO metrics (name, unit, value, timestamp) VALUES (?, ?, ?, ?)`)
	if err != nil {
		db.Close()
		return fmt.Errorf("prepare insert stmt: %w", err)
	}

	// Start the background writer goroutine
	s.writerWg.Add(1)
	go s.writerLoop()

	// Start the periodic daily retention prune
	s.pruneWg.Add(1)
	go s.dailyPruneLoop()

	globalStorage = s
	LogInfo("Persistent storage initialized at %s", path)
	return nil
}

// GetStorage returns the global storage instance.
func GetStorage() *Storage {
	return globalStorage
}

// Close closes the database connection after flushing pending writes.
func (s *Storage) Close() error {
	// Signal the daily prune loop to stop
	close(s.closeCh)
	s.pruneWg.Wait()

	// Signal the writer loop to flush and stop by closing the metrics channel
	close(s.metricsCh)
	s.writerWg.Wait()

	if s.insertStmt != nil {
		s.insertStmt.Close()
	}
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// ── Migrations ──────────────────────────────────────────────────────────────

func (s *Storage) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS metrics (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			name TEXT NOT NULL,
			unit TEXT,
			value REAL NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
			level TEXT NOT NULL,
			module TEXT,
			message TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_metrics_name_time ON metrics(name, timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_logs_time ON logs(timestamp)`,
		`CREATE TABLE IF NOT EXISTS events (
			id TEXT PRIMARY KEY,
			timestamp DATETIME NOT NULL,
			category TEXT NOT NULL,
			level TEXT NOT NULL,
			title TEXT NOT NULL,
			detail TEXT NOT NULL DEFAULT '',
			module TEXT NOT NULL DEFAULT '',
			related TEXT NOT NULL DEFAULT '',
			metadata TEXT NOT NULL DEFAULT ''
		)`,
		`CREATE INDEX IF NOT EXISTS idx_events_time ON events(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_events_cat ON events(category, timestamp)`,
		`CREATE TABLE IF NOT EXISTS alerts (
			id TEXT PRIMARY KEY,
			timestamp DATETIME NOT NULL,
			level TEXT NOT NULL,
			metric TEXT NOT NULL,
			message TEXT NOT NULL,
			value REAL NOT NULL,
			threshold REAL NOT NULL,
			resolved INTEGER NOT NULL DEFAULT 0,
			resolved_at DATETIME
		)`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_time ON alerts(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_alerts_metric ON alerts(metric)`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

// ── Background Writer ───────────────────────────────────────────────────────

// writerLoop drains the metrics channel and inserts batches on a ticker.
func (s *Storage) writerLoop() {
	defer s.writerWg.Done()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var batch []MetricWrite

	for {
		select {
		case w, ok := <-s.metricsCh:
			if !ok {
				// Channel closed — flush remaining and exit
				s.insertBatch(batch)
				return
			}
			batch = append(batch, w)
			if len(batch) >= 32 {
				s.insertBatch(batch)
				batch = nil
			}

		case f := <-s.flushCh:
			// Internal flush signal — drain all pending writes then flush
			batch = s.drainMetrics(batch)
			s.insertBatch(batch)
			batch = nil
			close(f)

		case <-ticker.C:
			if len(batch) > 0 {
				s.insertBatch(batch)
				batch = nil
			}
		}
	}
}

// drainMetrics reads all currently available items from metricsCh
// into the provided batch and returns the accumulated batch.
func (s *Storage) drainMetrics(batch []MetricWrite) []MetricWrite {
	for {
		select {
		case w, ok := <-s.metricsCh:
			if !ok {
				return batch
			}
			batch = append(batch, w)
		default:
			return batch
		}
	}
}

// insertBatch writes a batch of metrics inside a single transaction.
func (s *Storage) insertBatch(batch []MetricWrite) {
	if len(batch) == 0 {
		return
	}

	tx, err := s.db.Begin()
	if err != nil {
		LogInfo("Batch insert begin tx: %v", err)
		return
	}

	stmt := tx.Stmt(s.insertStmt)
	for _, m := range batch {
		if _, err := stmt.Exec(m.Name, m.Unit, m.Value, m.Time); err != nil {
			LogInfo("Batch insert metric: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		LogInfo("Batch insert commit: %v", err)
		tx.Rollback()
	}
}

// flushMetrics blocks until all pending metric writes are persisted.
// Intended for tests; accessible within the package.
func (s *Storage) flushMetrics() {
	done := make(chan struct{})
	select {
	case s.flushCh <- done:
		<-done
	case <-s.closeCh:
	}
}

// dailyPruneLoop runs Prune once per day until the storage is closed.
func (s *Storage) dailyPruneLoop() {
	defer s.pruneWg.Done()
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.Prune(7 * 24 * time.Hour)
		case <-s.closeCh:
			return
		}
	}
}

// ── Metrics CRUD ────────────────────────────────────────────────────────────

// InsertMetric enqueues a single metric value for asynchronous batch write.
func (s *Storage) InsertMetric(name, unit string, value float64) error {
	s.metricsCh <- MetricWrite{Name: name, Unit: unit, Value: value, Time: time.Now()}
	return nil
}

// GetMetricHistory retrieves historical values for a metric.
func (s *Storage) GetMetricHistory(name string, limit int) ([]float64, error) {
	query := `SELECT value FROM metrics WHERE name = ? ORDER BY timestamp DESC LIMIT ?`
	rows, err := s.db.Query(query, name, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var values []float64
	for rows.Next() {
		var v float64
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		values = append(values, v)
	}

	// Reverse to maintain chronological order
	for i, j := 0, len(values)-1; i < j; i, j = i+1, j-1 {
		values[i], values[j] = values[j], values[i]
	}

	return values, nil
}

// MetricDataPoint pairs a metric value with its timestamp.
type MetricDataPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}

// GetMetricHistoryWithTimestamps retrieves historical values with timestamps for a metric.
func (s *Storage) GetMetricHistoryWithTimestamps(name string, limit int) ([]MetricDataPoint, error) {
	query := `SELECT timestamp, value FROM metrics WHERE name = ? ORDER BY timestamp DESC LIMIT ?`
	rows, err := s.db.Query(query, name, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var points []MetricDataPoint
	for rows.Next() {
		var p MetricDataPoint
		if err := rows.Scan(&p.Timestamp, &p.Value); err != nil {
			return nil, err
		}
		points = append(points, p)
	}

	// Reverse to maintain chronological order
	for i, j := 0, len(points)-1; i < j; i, j = i+1, j-1 {
		points[i], points[j] = points[j], points[i]
	}

	if points == nil {
		points = []MetricDataPoint{}
	}
	return points, nil
}

// ── Logs CRUD ───────────────────────────────────────────────────────────────

// InsertLog persists a log entry.
func (s *Storage) InsertLog(level, module, message string) error {
	query := `INSERT INTO logs (level, module, message, timestamp) VALUES (?, ?, ?, ?)`
	_, err := s.db.Exec(query, level, module, message, time.Now())
	return err
}

// QueryLogs retrieves filtered logs from the database.
func (s *Storage) QueryLogs(level, search string, limit int) ([]LogEntryData, error) {
	query := `SELECT timestamp, level, module, message FROM logs WHERE 1=1`
	args := []interface{}{}

	if level != "" {
		query += ` AND level = ?`
		args = append(args, level)
	}
	if search != "" {
		query += ` AND (message LIKE ? OR module LIKE ?)`
		args = append(args, "%"+search+"%", "%"+search+"%")
	}

	query += ` ORDER BY timestamp DESC LIMIT ?`
	args = append(args, limit)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []LogEntryData
	for rows.Next() {
		var e LogEntryData
		var t time.Time
		if err := rows.Scan(&t, &e.Level, &e.Module, &e.Message); err != nil {
			return nil, err
		}
		e.Timestamp = t.Format("2006/01/02 15:04:05")
		entries = append(entries, e)
	}
	return entries, nil
}

// LogEntryData is a helper for querying logs.
type LogEntryData struct {
	Timestamp string
	Level     string
	Module    string
	Message   string
}

// SourceCount pairs a module name with its occurrence count.
type SourceCount struct {
	Source string
	Count  int
}

// TrendingLogError holds a frequently occurring error.
type TrendingLogError struct {
	Message  string
	Count    int
	LastSeen time.Time
}

// CountLogsAfter returns the number of log entries after the given time.
func (s *Storage) CountLogsAfter(t time.Time) int {
	var count int
	s.db.QueryRow(`SELECT COUNT(*) FROM logs WHERE timestamp >= ?`, t).Scan(&count)
	return count
}

// CountLogsByLevel returns the number of log entries with the given level.
func (s *Storage) CountLogsByLevel(level string) int {
	var count int
	s.db.QueryRow(`SELECT COUNT(*) FROM logs WHERE level = ?`, level).Scan(&count)
	return count
}

// TopLogSources returns the top N log sources by count.
func (s *Storage) TopLogSources(limit int) ([]SourceCount, error) {
	rows, err := s.db.Query(
		`SELECT module, COUNT(*) as cnt FROM logs WHERE module != '' GROUP BY module ORDER BY cnt DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []SourceCount
	for rows.Next() {
		var sc SourceCount
		if err := rows.Scan(&sc.Source, &sc.Count); err != nil {
			return nil, err
		}
		results = append(results, sc)
	}
	if results == nil {
		results = []SourceCount{}
	}
	return results, nil
}

// TrendingLogErrors returns the top N most frequent error messages.
func (s *Storage) TrendingLogErrors(limit int) ([]TrendingLogError, error) {
	rows, err := s.db.Query(
		`SELECT message, COUNT(*) as cnt, MAX(timestamp) as last_seen FROM logs WHERE level = 'ERROR' GROUP BY message ORDER BY cnt DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []TrendingLogError
	for rows.Next() {
		var te TrendingLogError
		if err := rows.Scan(&te.Message, &te.Count, &te.LastSeen); err != nil {
			return nil, err
		}
		results = append(results, te)
	}
	if results == nil {
		results = []TrendingLogError{}
	}
	return results, nil
}

// Prune removes data older than the given duration.
func (s *Storage) Prune(olderThan time.Duration) {
	cutoff := time.Now().Add(-olderThan)
	s.db.Exec(`DELETE FROM metrics WHERE timestamp < ?`, cutoff)
	s.db.Exec(`DELETE FROM logs WHERE timestamp < ?`, cutoff)
	s.db.Exec(`DELETE FROM events WHERE timestamp < ?`, cutoff)
	LogInfo("Retention policy applied: Pruned data older than %v", olderThan)
}

// ── Events CRUD ────────────────────────────────────────────────────────────────

// InsertEvent persists a timeline event.
func (s *Storage) InsertEvent(evt TimelineEvent) error {
	// Encode related IDs as comma-separated
	related := ""
	for i, r := range evt.Related {
		if i > 0 {
			related += ","
		}
		related += r
	}

	// Encode metadata as JSON
	metaJSON := ""
	if len(evt.Metadata) > 0 {
		if b, err := json.Marshal(evt.Metadata); err == nil {
			metaJSON = string(b)
		}
	}

	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO events (id, timestamp, category, level, title, detail, module, related, metadata) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		evt.ID, evt.Timestamp, string(evt.Category), evt.Level.String(), evt.Title, evt.Detail, evt.Module, related, metaJSON,
	)
	if err != nil {
		LogWarn("InsertEvent failed: %v", err)
	}
	return err
}

// QueryEvents retrieves events, newest first, with optional filters.
func (s *Storage) QueryEvents(category string, level string, limit int, offset int) ([]TimelineEvent, error) {
	query := `SELECT id, timestamp, category, level, title, detail, module, related, metadata FROM events WHERE 1=1`
	args := []interface{}{}

	if category != "" {
		query += ` AND category = ?`
		args = append(args, category)
	}
	if level != "" {
		query += ` AND level = ?`
		args = append(args, level)
	}

	query += ` ORDER BY timestamp DESC LIMIT ? OFFSET ?`
	args = append(args, limit, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []TimelineEvent
	for rows.Next() {
		var e TimelineEvent
		var cat, lvl, related, metaJSON string
		if err := rows.Scan(&e.ID, &e.Timestamp, &cat, &lvl, &e.Title, &e.Detail, &e.Module, &related, &metaJSON); err != nil {
			return nil, err
		}
		e.Category = EventCategory(cat)
		switch lvl {
		case "warning":
			e.Level = EventWarning
		case "critical":
			e.Level = EventCritical
		default:
			e.Level = EventInfo
		}
		// Parse related IDs
		if related != "" {
			e.Related = splitComma(related)
		}
		// Parse metadata JSON
		if metaJSON != "" {
			_ = json.Unmarshal([]byte(metaJSON), &e.Metadata)
		}
		events = append(events, e)
	}
	return events, nil
}

// GetEventByID retrieves a single event by ID.
func (s *Storage) GetEventByID(id string) (*TimelineEvent, error) {
	query := `SELECT id, timestamp, category, level, title, detail, module, related, metadata FROM events WHERE id = ?`
	row := s.db.QueryRow(query, id)

	var e TimelineEvent
	var cat, lvl, related, metaJSON string
	if err := row.Scan(&e.ID, &e.Timestamp, &cat, &lvl, &e.Title, &e.Detail, &e.Module, &related, &metaJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	e.Category = EventCategory(cat)
	switch lvl {
	case "warning":
		e.Level = EventWarning
	case "critical":
		e.Level = EventCritical
	default:
		e.Level = EventInfo
	}
	if related != "" {
		e.Related = splitComma(related)
	}
	if metaJSON != "" {
		_ = json.Unmarshal([]byte(metaJSON), &e.Metadata)
	}
	return &e, nil
}

// splitComma splits a comma-separated string, handling empty input.
func splitComma(s string) []string {
	if s == "" {
		return nil
	}
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}

// ── Alerts Persistence ─────────────────────────────────────────────────────

// AlertRecord represents a persisted alert.
type AlertRecord struct {
	ID         string
	Timestamp  time.Time
	Level      string
	Metric     string
	Message    string
	Value      float64
	Threshold  float64
	Resolved   bool
	ResolvedAt *time.Time
}

// InsertAlert persists a fired alert to SQLite.
func (s *Storage) InsertAlert(a AlertRecord) error {
	_, err := s.db.Exec(
		`INSERT OR REPLACE INTO alerts (id, timestamp, level, metric, message, value, threshold, resolved, resolved_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.Timestamp, a.Level, a.Metric, a.Message, a.Value, a.Threshold, boolToInt(a.Resolved), timePtrToSQL(a.ResolvedAt),
	)
	if err != nil {
		LogWarn("InsertAlert failed: %v", err)
	}
	return err
}

// UpdateAlertResolved marks an alert as resolved in SQLite.
func (s *Storage) UpdateAlertResolved(id string) error {
	_, err := s.db.Exec(
		`UPDATE alerts SET resolved = 1, resolved_at = ? WHERE id = ?`,
		time.Now(), id,
	)
	return err
}

// QueryAlertHistory returns all alerts, newest first, with an optional limit.
func (s *Storage) QueryAlertHistory(limit int) ([]AlertRecord, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := s.db.Query(
		`SELECT id, timestamp, level, metric, message, value, threshold, resolved, resolved_at FROM alerts ORDER BY timestamp DESC LIMIT ?`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []AlertRecord
	for rows.Next() {
		var a AlertRecord
		var resolved int
		var resolvedAt sql.NullTime
		if err := rows.Scan(&a.ID, &a.Timestamp, &a.Level, &a.Metric, &a.Message, &a.Value, &a.Threshold, &resolved, &resolvedAt); err != nil {
			return nil, err
		}
		a.Resolved = resolved != 0
		if resolvedAt.Valid {
			a.ResolvedAt = &resolvedAt.Time
		}
		alerts = append(alerts, a)
	}
	if alerts == nil {
		alerts = []AlertRecord{}
	}
	return alerts, nil
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func timePtrToSQL(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}
