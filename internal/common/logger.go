package common

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

const logChanCap = 1024

type logEntry struct {
	level string
	msg   string
}

var (
	logFile *os.File
	zlog    zerolog.Logger
	logWg   sync.WaitGroup
	logChan chan logEntry
	logOnce sync.Once
)

func startLogWorker() {
	logChan = make(chan logEntry, logChanCap)
	logWg.Add(1)
	go func() {
		defer logWg.Done()

		var batch []logEntry
		ticker := time.NewTicker(3 * time.Second)
		defer ticker.Stop()

		flush := func() {
			if len(batch) == 0 {
				return
			}
			s := GetStorage()
			if s != nil && !s.CheckClosed() {
				tx, err := s.Begin()
				if err != nil {
					return
				}
				query := `INSERT INTO logs (level, module, message, timestamp) VALUES (?, ?, ?, ?)`
				for _, entry := range batch {
					if _, err := tx.Exec(query, entry.level, "SYSTEM", entry.msg, time.Now().UTC().Format(time.RFC3339)); err != nil {
						continue
					}
				}
				_ = tx.Commit()
			}
			batch = nil
		}

		for {
			select {
			case entry, ok := <-logChan:
				if !ok {
					flush()
					return
				}
				batch = append(batch, entry)
				if len(batch) >= 50 {
					flush()
				}
			case <-ticker.C:
				flush()
			}
		}
	}()
}

// InitLogger initializes the session logger locally.
func InitLogger(filename string) error {
	if filename == "" {
		filename = filepath.Join("logs", "allopsfull.log")
	}

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	logFile = f

	// Configure zerolog with JSON output to file
	output := zerolog.MultiLevelWriter(
		zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339, NoColor: false},
		f,
	)

	zlog = zerolog.New(output).
		With().
		Timestamp().
		Str("app", "allopsfull").
		Logger()

	// AUDIT: Default to Debug level to capture Phase 2 telemetry
	if os.Getenv("GO_ENV") != "production" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	logOnce.Do(startLogWorker)

	LogInfo("Session started")
	return nil
}

// CloseLogger closes the log file and drains pending log writes.
func CloseLogger() {
	if logFile != nil {
		LogInfo("Session ended")
		// Close channel to signal worker, then wait for drain
		if logChan != nil {
			close(logChan)
		}
		logWg.Wait()
		logFile.Close()
	}
}

// enqueueLog sends a log entry to the worker channel, dropping if full.
func enqueueLog(level, msg string) {
	logOnce.Do(startLogWorker)
	select {
	case logChan <- logEntry{level: level, msg: msg}:
	default:
		// Channel full — drop to avoid blocking callers
	}
}

// LogInfo logs an informational message.
func LogInfo(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	zlog.Info().Msg(msg)
	enqueueLog("INFO", msg)
}

// LogWarn logs a warning message.
func LogWarn(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	zlog.Warn().Msg(msg)
	enqueueLog("WARN", msg)
}

// LogError logs an error message.
func LogError(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	zlog.Error().Msg(msg)
	enqueueLog("ERROR", msg)
}

// LogDebug logs a debug message (only in dev builds or when enabled).
func LogDebug(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	zlog.Debug().Msg(msg)
}

// SetLogLevel updates the dynamic log level of the zerolog instance.
func SetLogLevel(level string) {
	switch level {
	case "trace":
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}
}
