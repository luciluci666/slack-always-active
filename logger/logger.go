package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
	warnLogger  *log.Logger
	logFile     *os.File
)

// Init initializes the logger with the specified log file path
func Init(logPath string) error {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// Open log file
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	// Create multi-writer for both file and stdout
	multiWriter := io.MultiWriter(os.Stdout, file)

	// Initialize loggers
	infoLogger = log.New(multiWriter, "INFO: ", log.Ldate|log.Ltime)
	errorLogger = log.New(multiWriter, "ERROR: ", log.Ldate|log.Ltime)
	warnLogger = log.New(multiWriter, "WARN: ", log.Ldate|log.Ltime)

	logFile = file
	return nil
}

// Close closes the log file
func Close() {
	if logFile != nil {
		logFile.Close()
	}
}

// Info logs an info message
func Info(format string, v ...interface{}) {
	infoLogger.Printf(format, v...)
}

// Error logs an error message
func Error(format string, v ...interface{}) {
	errorLogger.Printf(format, v...)
}

// Warn logs a warning message
func Warn(format string, v ...interface{}) {
	warnLogger.Printf(format, v...)
}

// Printf logs a message with the default format
func Printf(format string, v ...interface{}) {
	infoLogger.Printf(format, v...)
}
