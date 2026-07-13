package common

import (
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

var (
	logFile *os.File
	zlog    zerolog.Logger
)

// InitLogger initializes the session logger with zerolog.
func InitLogger(filename string) error {
	if filename == "" {
		filename = "opsforall.log"
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
		Str("app", "opsforall").
		Logger()

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

// LogInfo logs an informational message.
func LogInfo(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	zlog.Info().Msg(msg)
	if s := GetStorage(); s != nil {
		go func() {
			defer RecoverPanic()
			s.InsertLog("INFO", "SYSTEM", msg)
		}()
	}
}

// LogWarn logs a warning message.
func LogWarn(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	zlog.Warn().Msg(msg)
	if s := GetStorage(); s != nil {
		go func() {
			defer RecoverPanic()
			s.InsertLog("WARN", "SYSTEM", msg)
		}()
	}
}

// LogError logs an error message.
func LogError(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	zlog.Error().Msg(msg)
	if s := GetStorage(); s != nil {
		go func() {
			defer RecoverPanic()
			s.InsertLog("ERROR", "SYSTEM", msg)
		}()
	}
}

// LogDebug logs a debug message (only in dev builds or when enabled).
func LogDebug(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	zlog.Debug().Msg(msg)
}
