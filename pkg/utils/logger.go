package utils

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logFile  *os.File
	logMutex sync.Mutex
	logPath  string
)

// InitLogger initializes the debug logger
func InitLogger() error {
	logMutex.Lock()
	defer logMutex.Unlock()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = "."
	}

	logDir := filepath.Join(homeDir, ".lazyazure")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath = filepath.Join(logDir, "debug.log")

	// Truncate on start for clean logs
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	logFile = f

	// Write initial log entries directly (don't use Log() since we hold the mutex)
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	logFile.WriteString(fmt.Sprintf("[%s] === LazyAzure Debug Log Started ===\n", timestamp))
	logFile.WriteString(fmt.Sprintf("[%s] Log file: %s\n", timestamp, logPath))
	logFile.Sync()

	return nil
}

// Log writes a log message with timestamp
func Log(format string, args ...interface{}) {
	logMutex.Lock()
	defer logMutex.Unlock()

	if logFile == nil {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	msg := fmt.Sprintf(format, args...)
	line := fmt.Sprintf("[%s] %s\n", timestamp, msg)

	logFile.WriteString(line)
	logFile.Sync() // Force flush to disk
}

// CloseLogger closes the log file
func CloseLogger() {
	logMutex.Lock()
	defer logMutex.Unlock()

	if logFile != nil {
		Log("=== Logger closed ===")
		logFile.Close()
		logFile = nil
	}
}

// GetLogPath returns the current log file path
func GetLogPath() string {
	return logPath
}
