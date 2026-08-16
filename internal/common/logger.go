package common

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
)

var (
	// loggerMu guards logFile and logger against concurrent re-init
	// (UpdateLogsConfig / Shutdown) while background goroutines log (M1).
	loggerMu sync.RWMutex
	logFile  *os.File
	logger   *slog.Logger
	logLevel = new(slog.LevelVar) // supports dynamic level changes at runtime
)

// defaultLogger is used before InitLogger runs (e.g. in tests or early boot).
// slog.Logger is nil until initialized, so a zero-value logger would panic on
// first use — this guarantees LogInfo/LogWarn/LogError/LogDebug are always safe.
var defaultLogger = slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

// getLogger returns the active logger, falling back to the discard default
// when InitLogger has not been called yet.
func getLogger() *slog.Logger {
	loggerMu.RLock()
	defer loggerMu.RUnlock()
	if logger != nil {
		return logger
	}
	return defaultLogger
}

// multiHandler fans out log records to multiple slog.Handlers.
type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(_ context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(context.Background(), level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: handlers}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: handlers}
}

// InitLogger initializes the session logger locally.
func InitLogger(filename string) error {
	if filename == "" {
		filename = filepath.Join("logs", "universalops.log")
	}

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	// Default to Debug in dev, Info in production
	if os.Getenv("GO_ENV") != "production" {
		logLevel.Set(slog.LevelDebug)
	} else {
		logLevel.Set(slog.LevelInfo)
	}

	// Console handler: human-readable text output to stderr
	consoleHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: logLevel,
	})

	// File handler: structured JSON output to file
	fileHandler := slog.NewJSONHandler(f, &slog.HandlerOptions{
		Level: logLevel,
	})

	// Fan-out to both handlers
	handler := &multiHandler{
		handlers: []slog.Handler{consoleHandler, fileHandler},
	}

	newLogger := slog.New(handler).With("app", "universalops")

	loggerMu.Lock()
	if logFile != nil && logFile != f {
		// Close a previous file if re-initializing without CloseLogger.
		logFile.Close()
	}
	logFile = f
	logger = newLogger
	loggerMu.Unlock()

	slog.SetDefault(newLogger)

	LogInfo("Session started")
	return nil
}

// CloseLogger closes the log file.
func CloseLogger() {
	loggerMu.Lock()
	defer loggerMu.Unlock()
	if logFile != nil {
		// Log via the captured logger directly — getLogger would re-acquire
		// the lock and deadlock while we hold the write lock.
		if logger != nil {
			logger.Info("Session ended")
		}
		logFile.Close()
		logFile = nil
	}
}

// enqueueLog sends a log entry to the persistent storage.
func enqueueLog(level, msg string) {
	s := GetStorage()
	if s != nil {
		_ = s.InsertLog(level, "SYSTEM", msg)
	}
}

// LogInfo logs an informational message.
func LogInfo(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	getLogger().Info(msg)
	enqueueLog("INFO", msg)
}

// LogWarn logs a warning message.
func LogWarn(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	getLogger().Warn(msg)
	enqueueLog("WARN", msg)
}

// LogError logs an error message.
func LogError(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	getLogger().Error(msg)
	enqueueLog("ERROR", msg)
}

// LogDebug logs a debug message (only in dev builds or when enabled).
func LogDebug(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	getLogger().Debug(msg)
}

// SetLogLevel updates the dynamic log level at runtime.
func SetLogLevel(level string) {
	switch level {
	case "trace":
		logLevel.Set(slog.LevelDebug - 4) // slog has no trace; use sub-debug
	case "debug":
		logLevel.Set(slog.LevelDebug)
	case "warn":
		logLevel.Set(slog.LevelWarn)
	case "error":
		logLevel.Set(slog.LevelError)
	default:
		logLevel.Set(slog.LevelInfo)
	}
}
