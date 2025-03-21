package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

var (
	fileLogger *log.Logger
	file       *os.File
)

// Init initializes the logger with a file
func Init(logPath string) error {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	// Open log file with append mode
	var err error
	file, err = os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	// Create multi-writer for both file and stdout
	multiWriter := io.MultiWriter(os.Stdout, file)

	// Initialize logger with timestamp
	fileLogger = log.New(multiWriter, "", log.LstdFlags)

	return nil
}

// Close closes the log file
func Close() error {
	if file != nil {
		return file.Close()
	}
	return nil
}

// Printf logs a formatted message
func Printf(format string, v ...interface{}) {
	if fileLogger != nil {
		fileLogger.Printf(format, v...)
	}
}

// Error logs an error message
func Error(format string, v ...interface{}) {
	if fileLogger != nil {
		fileLogger.Printf("ERROR: "+format, v...)
	}
}
