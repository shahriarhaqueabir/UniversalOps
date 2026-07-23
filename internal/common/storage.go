package common

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
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

// LogWrite represents a single log entry enqueued for asynchronous write.
type LogWrite struct {
	Level   string
	Module  string
	Message string
	Time    time.Time
}

// Storage manages the persistent SQLite database.
type Storage struct {
	db         *sql.DB
	insertStmt *sql.Stmt
	metricsCh  chan MetricWrite
	logsCh     chan LogWrite
	closeCh    chan struct{}
	writerWg   sync.WaitGroup
	pruneWg    sync.WaitGroup
	flushCh    chan chan struct{}

	// mu protects metricsCh and closed flag
	mu     sync.RWMutex
	closed bool

	// Instrumentation metrics
	writeCount        uint64
	totalWriteDur     time.Duration
	insertMetricCalls uint64
	statsMu           sync.Mutex
}

var (
	// DefaultDBName is the name of the SQLite database file.
	DefaultDBName = "universalops.db"

	// globalStorageMu serialises access to globalStorage so that
	// InitStorage (writer) and GetStorage (readers) never race.
	globalStorageMu sync.RWMutex
	globalStorage   *Storage
)

// InitStorage initializes the global SQLite storage.
func InitStorage(path string) error {
	LogDebug("SQLITE_METRICS | Initializing storage at %s", path)
	if path == "" {
		path = DefaultDBName
	}

	// Close any previous storage to prevent writerLoop leaks across tests
	if prev := GetStorage(); prev != nil {
		prev.Close()
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

	// Performance settings: 2 connections to avoid deadlock between the
	// background writer (writerLoop → insertBatch → Begin) and the log
	// worker (flush → Begin → tx.Exec). SQLite with WAL mode supports
	// concurrent reads and one writer.
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(2)

	// WAL mode and performance pragmas
	pragmas := []string{
		"PRAGMA journal_mode=WAL",
		"PRAGMA synchronous=NORMAL",
		"PRAGMA cache_size=-8000",
		"PRAGMA busy_timeout=5000",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			db.Close()
			return fmt.Errorf("pragma %s: %w", p, err)
		}
	}

	s := &Storage{
		db:        db,
		metricsCh: make(chan MetricWrite, 2048), // Increased from 256 to handle spikes
		logsCh:    make(chan LogWrite, 1024),
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

	LogDebug("SQLITE_METRICS | Initializing...")
	LogDebug("SQLITE_METRICS | starting background loops")

	// Start the background writer goroutine
	s.writerWg.Add(1)
	go s.writerLoop()

	// Start the periodic daily retention prune
	s.pruneWg.Add(1)
	go s.dailyPruneLoop()

	// Phase 2: Start background metrics logger
	go s.metricsLoggerLoop()

	globalStorageMu.Lock()
	globalStorage = s
	globalStorageMu.Unlock()
	LogInfo("Persistent storage initialized at %s", path)
	return nil
}

// GetStorage returns the global storage instance.
func GetStorage() *Storage {
	globalStorageMu.RLock()
	defer globalStorageMu.RUnlock()
	return globalStorage
}

// Close closes the database connection after flushing pending writes.
func (s *Storage) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	close(s.closeCh)
	close(s.metricsCh)
	s.mu.Unlock()

	s.pruneWg.Wait()
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

// Begin starts a new SQL transaction.
func (s *Storage) Begin() (*sql.Tx, error) {
	return s.db.Begin()
}

func (s *Storage) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS schema_versions (
			version INTEGER PRIMARY KEY,
			applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`INSERT OR IGNORE INTO schema_versions (version) VALUES (1)`,
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
		`CREATE TABLE IF NOT EXISTS conversations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_conv_session ON conversations(session_id, timestamp)`,
		`CREATE TABLE IF NOT EXISTS settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS forensics (
			id TEXT PRIMARY KEY,
			timestamp DATETIME NOT NULL,
			type TEXT NOT NULL,
			title TEXT NOT NULL,
			data_json TEXT NOT NULL,
			metadata TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS reports (
			id TEXT PRIMARY KEY,
			timestamp DATETIME NOT NULL,
			type TEXT NOT NULL,
			score INTEGER NOT NULL,
			data_json TEXT NOT NULL
		)`,
		`CREATE INDEX IF NOT EXISTS idx_reports_type_time ON reports(type, timestamp)`,
		`CREATE TABLE IF NOT EXISTS baselines (
			metric TEXT PRIMARY KEY,
			avg REAL NOT NULL,
			stddev REAL NOT NULL,
			last_updated DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS health_scores (
			day DATE PRIMARY KEY,
			score INTEGER NOT NULL,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS custom_workflows (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT NOT NULL,
			definition_json TEXT NOT NULL,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS incidents (
			id TEXT PRIMARY KEY,
			timestamp DATETIME NOT NULL,
			title TEXT NOT NULL,
			details TEXT NOT NULL,
			report_ids TEXT,
			severity TEXT NOT NULL
		)`,
	}

	for _, q := range queries {
		if _, err := s.db.Exec(q); err != nil {
			return err
		}
	}
	return nil
}

// ── Background Writer ───────────────────────────────────────────────────────

// writerLoop drains the metrics and logs channels and inserts batches on a ticker.
func (s *Storage) writerLoop() {
	defer RecoverPanic()
	defer s.writerWg.Done()
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	var metricsBatch []MetricWrite
	var logsBatch []LogWrite

	for {
		select {
		case w, ok := <-s.metricsCh:
			if !ok {
				s.insertMetricsBatch(metricsBatch)
				s.insertLogsBatch(logsBatch)
				return
			}
			metricsBatch = append(metricsBatch, w)
			// PERFORMANCE: Larger batches (512) and opportunistic draining
			// reduces transaction overhead significantly during spikes.
			if len(metricsBatch) >= 512 {
				metricsBatch = s.drainMetrics(metricsBatch)
				s.insertMetricsBatch(metricsBatch)
				metricsBatch = nil
			}

		case l, ok := <-s.logsCh:
			if !ok {
				// Handled by metricsCh close usually, but for safety:
				s.insertMetricsBatch(metricsBatch)
				s.insertLogsBatch(logsBatch)
				return
			}
			logsBatch = append(logsBatch, l)
			if len(logsBatch) >= 256 {
				logsBatch = s.drainLogs(logsBatch)
				s.insertLogsBatch(logsBatch)
				logsBatch = nil
			}

		case f := <-s.flushCh:
			// Internal flush signal — drain all pending writes then flush
			metricsBatch = s.drainMetrics(metricsBatch)
			logsBatch = s.drainLogs(logsBatch)
			s.insertMetricsBatch(metricsBatch)
			s.insertLogsBatch(logsBatch)
			metricsBatch = nil
			logsBatch = nil
			close(f)

		case <-ticker.C:
			// Ticker ensures data is persisted even if batches aren't full.
			if len(metricsBatch) > 0 {
				metricsBatch = s.drainMetrics(metricsBatch)
				s.insertMetricsBatch(metricsBatch)
				metricsBatch = nil
			}
			if len(logsBatch) > 0 {
				logsBatch = s.drainLogs(logsBatch)
				s.insertLogsBatch(logsBatch)
				logsBatch = nil
			}
		}
	}
}

// drainMetrics reads all currently available items from metricsCh
// into the provided batch and returns the accumulated batch.
// Limits to 2000 items to prevent unbounded transaction size.
func (s *Storage) drainMetrics(batch []MetricWrite) []MetricWrite {
	for len(batch) < 2000 {
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
	return batch
}

// drainLogs reads all currently available items from logsCh
// into the provided batch and returns the accumulated batch.
func (s *Storage) drainLogs(batch []LogWrite) []LogWrite {
	for len(batch) < 1000 {
		select {
		case l, ok := <-s.logsCh:
			if !ok {
				return batch
			}
			batch = append(batch, l)
		default:
			return batch
		}
	}
	return batch
}

func (s *Storage) metricsLoggerLoop() {
	ticker := time.NewTicker(5 * time.Second) // AUDIT: Reduced from 30s to 5s for fast verification
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// PROBE: Force a test write to see if it registers
			s.InsertMetric("probe.heartbeat", "bool", 1.0)

			s.statsMu.Lock()
			count := s.writeCount
			dur := s.totalWriteDur
			s.writeCount = 0
			s.totalWriteDur = 0
			s.statsMu.Unlock()

			if count > 0 {
				avg := dur / time.Duration(count)
				LogDebug("SQLITE_METRICS | batches=%d | total_dur=%v | avg_dur=%v", count, dur, avg)
			} else {
				LogDebug("SQLITE_METRICS | heartbeat (no writes)")
			}
		case <-s.closeCh:
			return
		}
	}
}

// insertMetricsBatch writes a batch of metrics inside a single transaction.
func (s *Storage) insertMetricsBatch(batch []MetricWrite) {
	if len(batch) == 0 {
		return
	}

	start := time.Now()
	tx, err := s.db.Begin()
	if err != nil {
		LogInfo("Batch insert metrics begin tx: %v", err)
		return
	}

	stmt := tx.Stmt(s.insertStmt)
	for _, m := range batch {
		if _, err := stmt.Exec(m.Name, m.Unit, m.Value, m.Time.UTC().Format(time.RFC3339)); err != nil {
			LogInfo("Batch insert metric: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		LogInfo("Batch insert metrics commit: %v", err)
		tx.Rollback()
		return
	}

	dur := time.Since(start)
	s.statsMu.Lock()
	s.writeCount++
	s.totalWriteDur += dur
	s.statsMu.Unlock()
}

// insertLogsBatch writes a batch of logs inside a single transaction.
func (s *Storage) insertLogsBatch(batch []LogWrite) {
	if len(batch) == 0 {
		return
	}

	tx, err := s.db.Begin()
	if err != nil {
		LogInfo("Batch insert logs begin tx: %v", err)
		return
	}

	query := `INSERT INTO logs (level, module, message, timestamp) VALUES (?, ?, ?, ?)`
	for _, l := range batch {
		if _, err := tx.Exec(query, l.Level, l.Module, l.Message, l.Time.UTC().Format(time.RFC3339)); err != nil {
			LogInfo("Batch insert log: %v", err)
		}
	}

	if err := tx.Commit(); err != nil {
		LogInfo("Batch insert logs commit: %v", err)
		tx.Rollback()
		return
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
	defer RecoverPanic()
	defer s.pruneWg.Done()
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Weekly maintenance ticker for Vacuum
	maintTicker := time.NewTicker(7 * 24 * time.Hour)
	defer maintTicker.Stop()

	for {
		select {
		case <-ticker.C:
			// Configurable retention (default 7 days)
			retentionDays := 7
			if val, err := s.GetSetting("retention_days"); err == nil && val != "" {
				if d, err := strconv.Atoi(val); err == nil && d > 0 {
					retentionDays = d
				}
			}
			s.Prune(time.Duration(retentionDays) * 24 * time.Hour)

		case <-maintTicker.C:
			// Automated maintenance: Vacuum if fragmented > 25%
			if frag := s.GetFragmentation(); frag > 25 {
				LogInfo("Storage: High fragmentation detected (%.1f%%). Running Vacuum.", frag)
				_ = s.Vacuum()
			}

		case <-s.closeCh:
			return
		}
	}
}

// Vacuum rebuilds the database to reclaim space.
func (s *Storage) Vacuum() error {
	_, err := s.db.Exec(`VACUUM`)
	return err
}

// Analyze updates query statistics.
func (s *Storage) Analyze() error {
	_, err := s.db.Exec(`ANALYZE`)
	return err
}

// GetFragmentation calculates the percentage of free pages in the database.
func (s *Storage) GetFragmentation() float64 {
	var freeCount, pageCount int
	_ = s.db.QueryRow(`PRAGMA freelist_count`).Scan(&freeCount)
	_ = s.db.QueryRow(`PRAGMA page_count`).Scan(&pageCount)
	if pageCount == 0 {
		return 0
	}
	return (float64(freeCount) / float64(pageCount)) * 100
}

// ── Metrics CRUD ────────────────────────────────────────────────────────────

// InsertMetric enqueues a single metric value for asynchronous batch write.
// M3: use a non-blocking send with a timeout to prevent the caller from
// stalling indefinitely if the writer loop is backed up. Dropped samples are
// logged at most once per 10-second window to avoid log spam.
var (
	lastDropLog   time.Time
	lastDropLogMu sync.Mutex
)

func (s *Storage) InsertMetric(name, unit string, value float64) error {
	atomic.AddUint64(&s.insertMetricCalls, 1)
	LogDebug("SQLITE_METRICS | InsertMetric called: %s", name)
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return fmt.Errorf("storage closed")
	}

	w := MetricWrite{Name: name, Unit: unit, Value: value, Time: time.Now()}

	// PERFORMANCE FIX: Reduced timeout from 500ms to 10ms.
	// A monitoring app should NEVER block the caller (like the UI or high-freq collector)
	// if the background persistence layer is backed up. Data freshness > History.
	timer := time.NewTimer(10 * time.Millisecond)
	defer timer.Stop()

	select {
	case s.metricsCh <- w:
		return nil
	case <-timer.C:
		// Channel full and timeout reached — drop sample and log a warning
		lastDropLogMu.Lock()
		if time.Since(lastDropLog) > 10*time.Second {
			lastDropLog = time.Now()
			LogWarn("InsertMetric: metrics queue backed up, dropping sample (name=%s). WORKSTATION DISK I/O IS SATURATED.", name)
		}
		lastDropLogMu.Unlock()
		return fmt.Errorf("queue full")
	}
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
		var tsRaw string
		if err := rows.Scan(&tsRaw, &p.Value); err != nil {
			return nil, err
		}
		// Parse timestamp from SQLite (usually ISO8601 or YYYY-MM-DD HH:MM:SS)
		if t, err := time.Parse("2006-01-02 15:04:05", tsRaw); err == nil {
			p.Timestamp = t
		} else if t, err := time.Parse(time.RFC3339, tsRaw); err == nil {
			p.Timestamp = t
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

// InsertLog enqueues a log entry for asynchronous batch write.
func (s *Storage) InsertLog(level, module, message string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return fmt.Errorf("storage closed")
	}

	l := LogWrite{
		Level:   level,
		Module:  module,
		Message: message,
		Time:    time.Now(),
	}

	select {
	case s.logsCh <- l:
		return nil
	default:
		// Channel full — drop log to avoid blocking
		return fmt.Errorf("logs queue full")
	}
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
		var tsRaw string
		var module sql.NullString
		if err := rows.Scan(&tsRaw, &e.Level, &module, &e.Message); err != nil {
			return nil, err
		}
		if module.Valid {
			e.Module = module.String
		}
		var t time.Time
		if parsed, err := time.Parse("2006-01-02 15:04:05", tsRaw); err == nil {
			t = parsed
		} else if parsed, err := time.Parse(time.RFC3339, tsRaw); err == nil {
			t = parsed
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
		var source sql.NullString
		if err := rows.Scan(&source, &sc.Count); err != nil {
			return nil, err
		}
		if source.Valid {
			sc.Source = source.String
		}
		results = append(results, sc)
	}
	if results == nil {
		results = []SourceCount{}
	}
	return results, nil
}

// CountLogsByLevelInRange returns the count of logs with the given level within a time range.
func (s *Storage) CountLogsByLevelInRange(level string, from, to time.Time) int {
	var count int
	s.db.QueryRow(`SELECT COUNT(*) FROM logs WHERE level = ? AND timestamp >= ? AND timestamp < ?`, level, from, to).Scan(&count)
	return count
}

// LogTimelineBucket holds a time-bucketed log count row.
type LogTimelineBucket struct {
	Bucket   string
	Total    int
	Errors   int
	Warnings int
	Info     int
}

// QueryLogTimeline returns time-bucketed log counts for charting.
// hours >= 24 buckets by hour; hours < 24 buckets by 5 minutes.
func (s *Storage) QueryLogTimeline(hours int) ([]LogTimelineBucket, error) {
	var bucketExpr string
	if hours >= 24 {
		bucketExpr = `strftime('%Y-%m-%d %H:00:00', timestamp)`
	} else {
		// 5-minute buckets via epoch rounding
		bucketExpr = `datetime(CAST(CAST(strftime('%s', timestamp) AS INTEGER) / 300 AS INTEGER) * 300, 'unixepoch')`
	}

	timeArg := fmt.Sprintf(`-%d hours`, hours)
	query := fmt.Sprintf(`
		SELECT
			%s as bucket,
			COUNT(*) as total,
			SUM(CASE WHEN level = 'ERROR' THEN 1 ELSE 0 END) as errors,
			SUM(CASE WHEN level = 'WARN' THEN 1 ELSE 0 END) as warnings,
			SUM(CASE WHEN level = 'INFO' THEN 1 ELSE 0 END) as info
		FROM logs
		WHERE timestamp > datetime('now', ?)
		GROUP BY bucket
		ORDER BY bucket ASC
	`, bucketExpr)

	rows, err := s.db.Query(query, timeArg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []LogTimelineBucket
	for rows.Next() {
		var b LogTimelineBucket
		var bucket *string
		var errors, warnings, info sql.NullInt64
		if err := rows.Scan(&bucket, &b.Total, &errors, &warnings, &info); err != nil {
			return nil, fmt.Errorf("scan timeline: %w", err)
		}
		if bucket != nil {
			b.Bucket = *bucket
		} else {
			continue // skip NULL buckets
		}
		b.Errors = int(errors.Int64)
		b.Warnings = int(warnings.Int64)
		b.Info = int(info.Int64)
		results = append(results, b)
	}
	if results == nil {
		results = []LogTimelineBucket{}
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
		var tsRaw string
		if err := rows.Scan(&te.Message, &te.Count, &tsRaw); err != nil {
			return nil, err
		}
		if t, err := time.Parse("2006-01-02 15:04:05", tsRaw); err == nil {
			te.LastSeen = t
		} else if t, err := time.Parse(time.RFC3339, tsRaw); err == nil {
			te.LastSeen = t
		}
		results = append(results, te)
	}
	if results == nil {
		results = []TrendingLogError{}
	}
	return results, nil
}

// Prune removes data older than the given duration. Covers every
// timestamped table, including alerts and conversations which were
// previously left out and grew unbounded (M5).
func (s *Storage) Prune(olderThan time.Duration) {
	cutoff := time.Now().Add(-olderThan)
	tx, err := s.db.Begin()
	if err != nil {
		LogWarn("Prune tx begin failed: %v", err)
		return
	}
	defer tx.Rollback()

	tx.Exec(`DELETE FROM metrics WHERE timestamp < ?`, cutoff)
	tx.Exec(`DELETE FROM logs WHERE timestamp < ?`, cutoff)
	tx.Exec(`DELETE FROM events WHERE timestamp < ?`, cutoff)
	tx.Exec(`DELETE FROM alerts WHERE timestamp < ? AND resolved = 1`, cutoff)
	tx.Exec(`DELETE FROM conversations WHERE timestamp < ?`, cutoff)
	tx.Exec(`DELETE FROM forensics WHERE timestamp < ?`, cutoff)
	tx.Exec(`DELETE FROM reports WHERE timestamp < ?`, cutoff)
	tx.Exec(`DELETE FROM incidents WHERE timestamp < ?`, cutoff)
	tx.Exec(`DELETE FROM health_scores WHERE timestamp < ?`, cutoff)

	if err := tx.Commit(); err != nil {
		LogWarn("Prune tx commit failed: %v", err)
		return
	}
	LogInfo("Retention policy applied: Pruned data older than %v", olderThan)
}

// ── Events CRUD ────────────────────────────────────────────────────────────────

// InsertEvent persists a timeline event.
// NOTE: mu is NOT held across the DB call to prevent deadlock with Close().
func (s *Storage) InsertEvent(evt TimelineEvent, tx *sql.Tx) error {
	s.mu.RLock()
	closed := s.closed
	s.mu.RUnlock()
	if closed && tx == nil {
		return fmt.Errorf("storage closed")
	}
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

	_, err := s.getDB(tx).Exec(
		`INSERT OR IGNORE INTO events (id, timestamp, category, level, title, detail, module, related, metadata) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		evt.ID, evt.Timestamp.UTC().Format(time.RFC3339), string(evt.Category), evt.Level.String(), evt.Title, evt.Detail, evt.Module, related, metaJSON,
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
		var tsRaw, cat, lvl, related, metaJSON string
		if err := rows.Scan(&e.ID, &tsRaw, &cat, &lvl, &e.Title, &e.Detail, &e.Module, &related, &metaJSON); err != nil {
			return nil, err
		}
		if t, err := time.Parse("2006-01-02 15:04:05", tsRaw); err == nil {
			e.Timestamp = t
		} else if t, err := time.Parse(time.RFC3339, tsRaw); err == nil {
			e.Timestamp = t
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
	var tsRaw, cat, lvl, related, metaJSON string
	if err := row.Scan(&e.ID, &tsRaw, &cat, &lvl, &e.Title, &e.Detail, &e.Module, &related, &metaJSON); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if t, err := time.Parse("2006-01-02 15:04:05", tsRaw); err == nil {
		e.Timestamp = t
	} else if t, err := time.Parse(time.RFC3339, tsRaw); err == nil {
		e.Timestamp = t
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

// queryable is an interface that works with both sql.DB and sql.Tx.
type queryable interface {
	Exec(query string, args ...interface{}) (sql.Result, error)
	Query(query string, args ...interface{}) (*sql.Rows, error)
	QueryRow(query string, args ...interface{}) *sql.Row
}

func (s *Storage) getDB(tx *sql.Tx) queryable {
	if tx != nil {
		return tx
	}
	return s.db
}

// CheckClosed returns true if the storage is closed.
func (s *Storage) CheckClosed() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.closed
}

// InsertAlert persists a fired alert to SQLite.
// NOTE: mu is NOT held across the DB call to prevent deadlock with Close().
func (s *Storage) InsertAlert(a AlertRecord, tx *sql.Tx) error {
	s.mu.RLock()
	closed := s.closed
	s.mu.RUnlock()
	if closed && tx == nil {
		return fmt.Errorf("storage closed")
	}

	resolvedAt := interface{}(nil)
	if a.ResolvedAt != nil {
		resolvedAt = a.ResolvedAt.UTC().Format(time.RFC3339)
	}

	_, err := s.getDB(tx).Exec(
		`INSERT OR REPLACE INTO alerts (id, timestamp, level, metric, message, value, threshold, resolved, resolved_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		a.ID, a.Timestamp.UTC().Format(time.RFC3339), a.Level, a.Metric, a.Message, a.Value, a.Threshold, boolToInt(a.Resolved), resolvedAt,
	)
	if err != nil {
		LogWarn("InsertAlert failed: %v", err)
	}
	return err
}

// UpdateAlertResolved marks an alert as resolved in SQLite.
// NOTE: mu is NOT held across the DB call to prevent deadlock with Close().
func (s *Storage) UpdateAlertResolved(id string, tx *sql.Tx) error {
	s.mu.RLock()
	closed := s.closed
	s.mu.RUnlock()
	if closed && tx == nil {
		return fmt.Errorf("storage closed")
	}
	_, err := s.getDB(tx).Exec(
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
		var tsRaw string
		var resolved int
		var resolvedAt sql.NullTime
		if err := rows.Scan(&a.ID, &tsRaw, &a.Level, &a.Metric, &a.Message, &a.Value, &a.Threshold, &resolved, &resolvedAt); err != nil {
			return nil, err
		}
		if t, err := time.Parse("2006-01-02 15:04:05", tsRaw); err == nil {
			a.Timestamp = t
		} else if t, err := time.Parse(time.RFC3339, tsRaw); err == nil {
			a.Timestamp = t
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

// ── Conversations ───────────────────────────────────────────────────────────

// ConversationMessage represents a single chat message.
type ConversationMessage struct {
	ID        int64     `json:"id"`
	SessionID string    `json:"session_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// InsertMessage persists a chat message to the conversations table.
func (s *Storage) InsertMessage(sessionID, role, content string) error {
	_, err := s.db.Exec(
		`INSERT INTO conversations (session_id, role, content, timestamp) VALUES (?, ?, ?, ?)`,
		sessionID, role, content, time.Now().UTC().Format(time.RFC3339),
	)
	return err
}

// GetMessages returns all messages for a given session, ordered by time.
func (s *Storage) GetMessages(sessionID string) ([]ConversationMessage, error) {
	rows, err := s.db.Query(
		`SELECT id, session_id, role, content, timestamp FROM conversations WHERE session_id = ? ORDER BY id ASC`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []ConversationMessage
	for rows.Next() {
		var m ConversationMessage
		var tsRaw string
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &tsRaw); err != nil {
			return nil, err
		}
		if t, err := time.Parse("2006-01-02 15:04:05", tsRaw); err == nil {
			m.Timestamp = t
		} else if t, err := time.Parse(time.RFC3339, tsRaw); err == nil {
			m.Timestamp = t
		}
		msgs = append(msgs, m)
	}
	if msgs == nil {
		msgs = []ConversationMessage{}
	}
	return msgs, nil
}

// ListSessions returns distinct session IDs with their latest timestamp and message count.
func (s *Storage) ListSessions() ([]map[string]interface{}, error) {
	rows, err := s.db.Query(
		`SELECT session_id, MAX(timestamp) as last_active, COUNT(*) as msg_count
		 FROM conversations GROUP BY session_id ORDER BY last_active DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []map[string]interface{}
	for rows.Next() {
		var sid string
		var lastActiveStr *string
		var count int
		if err := rows.Scan(&sid, &lastActiveStr, &count); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		var lastActive time.Time
		if lastActiveStr != nil {
			if t, err := time.Parse("2006-01-02 15:04:05", *lastActiveStr); err == nil {
				lastActive = t
			} else if t, err := time.Parse(time.RFC3339, *lastActiveStr); err == nil {
				lastActive = t
			}
		}
		sessions = append(sessions, map[string]interface{}{
			"session_id":  sid,
			"last_active": lastActive,
			"msg_count":   count,
		})
	}
	if sessions == nil {
		sessions = []map[string]interface{}{}
	}
	return sessions, nil
}

// DeleteSession removes all messages for a given session.
func (s *Storage) DeleteSession(sessionID string) error {
	_, err := s.db.Exec(`DELETE FROM conversations WHERE session_id = ?`, sessionID)
	return err
}

// ── Settings Persistence ───────────────────────────────────────────────────

// UpsertSetting inserts or updates a key-value setting.
func (s *Storage) UpsertSetting(key, value string) error {
	_, err := s.db.Exec(
		`INSERT INTO settings (key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP`,
		key, value,
	)
	return err
}

// ── Forensics Persistence ──────────────────────────────────────────────────

// ForensicRecord holds a structured forensic snapshot.
type ForensicRecord struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"`
	Title     string `json:"title"`
	DataJSON  string `json:"data_json"`
	Metadata  string `json:"metadata"`
}

// InsertForensic stores a new forensic snapshot.
func (s *Storage) InsertForensic(r ForensicRecord) error {
	_, err := s.db.Exec(
		`INSERT INTO forensics (id, timestamp, type, title, data_json, metadata) VALUES (?, ?, ?, ?, ?, ?)`,
		r.ID, r.Timestamp, r.Type, r.Title, r.DataJSON, r.Metadata,
	)
	return err
}

// ListForensics returns a summary list of all snapshots.
func (s *Storage) ListForensics() ([]ForensicRecord, error) {
	rows, err := s.db.Query(`SELECT id, timestamp, type, title FROM forensics ORDER BY timestamp DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []ForensicRecord
	for rows.Next() {
		var r ForensicRecord
		if err := rows.Scan(&r.ID, &r.Timestamp, &r.Type, &r.Title); err != nil {
			return nil, err
		}
		res = append(res, r)
	}
	return res, nil
}

// GetForensic retrieves a full snapshot by ID.
func (s *Storage) GetForensic(id string) (*ForensicRecord, error) {
	var r ForensicRecord
	err := s.db.QueryRow(`SELECT id, timestamp, type, title, data_json, metadata FROM forensics WHERE id = ?`, id).
		Scan(&r.ID, &r.Timestamp, &r.Type, &r.Title, &r.DataJSON, &r.Metadata)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &r, err
}

// ── Reports Persistence ────────────────────────────────────────────────────

// ReportRecord holds a persisted diagnostic or security report.
type ReportRecord struct {
	ID        string `json:"id"`
	Timestamp string `json:"timestamp"`
	Type      string `json:"type"` // "health", "security"
	Score     int    `json:"score"`
	DataJSON  string `json:"data_json"`
}

// InsertReport stores a new diagnostic report.
func (s *Storage) InsertReport(r ReportRecord) error {
	_, err := s.db.Exec(
		`INSERT INTO reports (id, timestamp, type, score, data_json) VALUES (?, ?, ?, ?, ?)`,
		r.ID, r.Timestamp, r.Type, r.Score, r.DataJSON,
	)
	return err
}

// ListReportsByType returns all reports of a specific type.
func (s *Storage) ListReportsByType(reportType string) ([]ReportRecord, error) {
	rows, err := s.db.Query(`SELECT id, timestamp, type, score FROM reports WHERE type = ? ORDER BY timestamp DESC`, reportType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []ReportRecord
	for rows.Next() {
		var r ReportRecord
		if err := rows.Scan(&r.ID, &r.Timestamp, &r.Type, &r.Score); err != nil {
			return nil, err
		}
		res = append(res, r)
	}
	return res, nil
}

// ListAllReports returns all reports across all types, newest first.
func (s *Storage) ListAllReports() ([]ReportRecord, error) {
	rows, err := s.db.Query(`SELECT id, timestamp, type, score FROM reports ORDER BY timestamp DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []ReportRecord
	for rows.Next() {
		var r ReportRecord
		if err := rows.Scan(&r.ID, &r.Timestamp, &r.Type, &r.Score); err != nil {
			return nil, err
		}
		res = append(res, r)
	}
	return res, nil
}

// GetReport retrieves a full report by ID.
func (s *Storage) GetReport(id string) (*ReportRecord, error) {
	var r ReportRecord
	err := s.db.QueryRow(`SELECT id, timestamp, type, score, data_json FROM reports WHERE id = ?`, id).
		Scan(&r.ID, &r.Timestamp, &r.Type, &r.Score, &r.DataJSON)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &r, err
}

// DeleteReport removes a report by ID.
func (s *Storage) DeleteReport(id string) error {
	_, err := s.db.Exec(`DELETE FROM reports WHERE id = ?`, id)
	return err
}

// ── Baselines Persistence ──────────────────────────────────────────────────

// BaselineEntry holds the statistical ground truth for a metric.
type BaselineEntry struct {
	Metric      string    `json:"metric"`
	Avg         float64   `json:"avg"`
	StdDev      float64   `json:"stddev"`
	LastUpdated time.Time `json:"last_updated"`
}

// UpsertBaseline updates or inserts a baseline record.
func (s *Storage) UpsertBaseline(b BaselineEntry) error {
	_, err := s.db.Exec(
		`INSERT INTO baselines (metric, avg, stddev, last_updated) VALUES (?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(metric) DO UPDATE SET avg = excluded.avg, stddev = excluded.stddev, last_updated = CURRENT_TIMESTAMP`,
		b.Metric, b.Avg, b.StdDev,
	)
	return err
}

// GetBaseline retrieves the baseline for a metric.
func (s *Storage) GetBaseline(metric string) (*BaselineEntry, error) {
	var b BaselineEntry
	err := s.db.QueryRow(`SELECT metric, avg, stddev, last_updated FROM baselines WHERE metric = ?`, metric).
		Scan(&b.Metric, &b.Avg, &b.StdDev, &b.LastUpdated)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &b, err
}

// ── Health Scorecard Persistence ──────────────────────────────────────────

// UpsertHealthScore stores the health score for the current day.
func (s *Storage) UpsertHealthScore(score int) error {
	day := time.Now().Format("2006-01-02")
	_, err := s.db.Exec(
		`INSERT INTO health_scores (day, score, timestamp) VALUES (?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(day) DO UPDATE SET score = excluded.score, timestamp = CURRENT_TIMESTAMP`,
		day, score,
	)
	return err
}

// GetHealthScoreTrend returns scores for the last N days.
func (s *Storage) GetHealthScoreTrend(days int) (map[string]int, error) {
	rows, err := s.db.Query(`SELECT day, score FROM health_scores ORDER BY day DESC LIMIT ?`, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[string]int)
	for rows.Next() {
		var day string
		var score int
		if err := rows.Scan(&day, &score); err != nil {
			return nil, err
		}
		res[day] = score
	}
	return res, nil
}

// ── Custom Workflows Persistence ──────────────────────────────────────────

// UpsertCustomWorkflow saves or updates a user-defined workflow.
func (s *Storage) UpsertCustomWorkflow(id, name, desc, jsonDef string) error {
	_, err := s.db.Exec(
		`INSERT INTO custom_workflows (id, name, description, definition_json, updated_at) VALUES (?, ?, ?, ?, CURRENT_TIMESTAMP)
		 ON CONFLICT(id) DO UPDATE SET name=excluded.name, description=excluded.description, definition_json=excluded.definition_json, updated_at=CURRENT_TIMESTAMP`,
		id, name, desc, jsonDef,
	)
	return err
}

// ListCustomWorkflows returns all user-defined workflows.
func (s *Storage) ListCustomWorkflows() ([]map[string]string, error) {
	rows, err := s.db.Query(`SELECT id, name, description, definition_json FROM custom_workflows ORDER BY updated_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []map[string]string
	for rows.Next() {
		var id, name, desc, def string
		if err := rows.Scan(&id, &name, &desc, &def); err != nil {
			return nil, err
		}
		res = append(res, map[string]string{"id": id, "name": name, "description": desc, "definition": def})
	}
	return res, nil
}

// DeleteCustomWorkflow removes a workflow from the database.
func (s *Storage) DeleteCustomWorkflow(id string) error {
	_, err := s.db.Exec(`DELETE FROM custom_workflows WHERE id = ?`, id)
	return err
}

// ── Incidents Persistence ──────────────────────────────────────────────────

// IncidentRecord holds a consolidated record of correlated alerts and diagnostics.
type IncidentRecord struct {
	ID        string   `json:"id"`
	Timestamp string   `json:"timestamp"`
	Title     string   `json:"title"`
	Details   string   `json:"details"`
	ReportIDs []string `json:"report_ids"`
	Severity  string   `json:"severity"`
}

// InsertIncident stores a new incident record.
func (s *Storage) InsertIncident(r IncidentRecord) error {
	reports := strings.Join(r.ReportIDs, ",")
	_, err := s.db.Exec(
		`INSERT INTO incidents (id, timestamp, title, details, report_ids, severity) VALUES (?, ?, ?, ?, ?, ?)`,
		r.ID, r.Timestamp, r.Title, r.Details, reports, r.Severity,
	)
	return err
}

// ListIncidents returns a summary of all incidents.
func (s *Storage) ListIncidents() ([]IncidentRecord, error) {
	rows, err := s.db.Query(`SELECT id, timestamp, title, severity FROM incidents ORDER BY timestamp DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []IncidentRecord
	for rows.Next() {
		var r IncidentRecord
		if err := rows.Scan(&r.ID, &r.Timestamp, &r.Title, &r.Severity); err != nil {
			return nil, err
		}
		res = append(res, r)
	}
	return res, nil
}

// GetSetting retrieves a setting value by key.
func (s *Storage) GetSetting(key string) (string, error) {
	var value string
	err := s.db.QueryRow(`SELECT value FROM settings WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

// ListSettings returns all persisted settings.
func (s *Storage) ListSettings() (map[string]string, error) {
	rows, err := s.db.Query(`SELECT key, value FROM settings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := make(map[string]string)
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return nil, err
		}
		res[k] = v
	}
	return res, nil
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
