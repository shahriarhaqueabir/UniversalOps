package common

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	logFile *os.File
	logger  *log.Logger
)

// InitLogger initializes the session logger.
func InitLogger(filename string) error {
	if filename == "" {
		filename = "hawkward.log"
	}

	f, err := os.OpenFile(filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open log file: %w", err)
	}

	logFile = f
	logger = log.New(f, "[HAWKWARD] ", log.LstdFlags)

	LogInfo("Session started at %s", time.Now().Format(time.RFC3339))
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
	if logger != nil {
		logger.Printf("[INFO] "+format, v...)
	}
}

// LogWarn logs a warning message.
func LogWarn(format string, v ...interface{}) {
	if logger != nil {
		logger.Printf("[WARN] "+format, v...)
	}
}

// LogError logs an error message.
func LogError(format string, v ...interface{}) {
	if logger != nil {
		logger.Printf("[ERROR] "+format, v...)
	}
}
