package common

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
)

var (
	logFile *os.File
	zlog    zerolog.Logger
)

// InitLogger initializes the session logger locally.
func InitLogger(filename string) error {
	if filename == "" {
		filename = filepath.Join("logs", "universalops.log")
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
		Str("app", "universalops").
		Logger()

	// AUDIT: Default to Debug level to capture Phase 2 telemetry
	if os.Getenv("GO_ENV") != "production" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	LogInfo("Session started")
	return nil
}

// CloseLogger closes the log file.
func CloseLogger() {
	if logFile != nil {
		LogInfo("Session ended")
		logFile.Close()
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
