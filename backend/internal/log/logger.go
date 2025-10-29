package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/sirupsen/logrus"
)

var (
	// globalLogger is the singleton logger instance
	globalLogger *logrus.Logger
)

const (
	// LogFile is the name of the log file
	LogFile = "app.log"
)

var (
	// LogDir is the directory where log files are stored
	// Uses local "logs" directory for cross-platform compatibility
	LogDir = getLogDir()
)

// getLogDir returns the log directory path
// Prefers /var/log/featureplus for Linux/Docker, falls back to ./logs for Windows/local dev
func getLogDir() string {
	// Check if we're in a Docker/Linux environment
	if runtime.GOOS == "linux" {
		if _, err := os.Stat("/var/log"); err == nil {
			return "/var/log/featureplus"
		}
	}
	
	// For Windows and local development, use absolute path
	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		// Fallback to relative path if we can't get cwd
		return "logs"
	}
	
	// Use absolute path to logs directory
	return filepath.Join(cwd, "logs")
}

// init initializes the global logger on package import
func init() {
	globalLogger = logrus.New()
	
	// Set JSON formatter for structured logging
	globalLogger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		PrettyPrint:     false,
	})
	
	// Set log level (can be configured via environment variable)
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "debug":
		globalLogger.SetLevel(logrus.DebugLevel)
	case "info":
		globalLogger.SetLevel(logrus.InfoLevel)
	case "warn":
		globalLogger.SetLevel(logrus.WarnLevel)
	case "error":
		globalLogger.SetLevel(logrus.ErrorLevel)
	default:
		globalLogger.SetLevel(logrus.InfoLevel)
	}
	
	// Create log directory if it doesn't exist
	if err := os.MkdirAll(LogDir, 0755); err != nil {
		// If we can't create the directory, log to stderr and continue with stdout only
		fmt.Fprintf(os.Stderr, "[LOGGER] Failed to create log directory %s: %v\n", LogDir, err)
		fmt.Fprintf(os.Stderr, "[LOGGER] Logging to stdout only\n")
		globalLogger.SetOutput(os.Stdout)
		return
	} else {
		fmt.Fprintf(os.Stderr, "[LOGGER] Log directory created/verified: %s\n", LogDir)
	}
	
	// Open log file
	logPath := filepath.Join(LogDir, LogFile)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		// If we can't open the log file, log to stderr and continue with stdout only
		fmt.Fprintf(os.Stderr, "[LOGGER] Failed to open log file %s: %v\n", logPath, err)
		fmt.Fprintf(os.Stderr, "[LOGGER] Logging to stdout only\n")
		globalLogger.SetOutput(os.Stdout)
		return
	}
	fmt.Fprintf(os.Stderr, "[LOGGER] Log file opened: %s\n", logPath)
	
	// Write to both stdout and file
	multiWriter := io.MultiWriter(os.Stdout, logFile)
	globalLogger.SetOutput(multiWriter)
	
	globalLogger.WithFields(logrus.Fields{
		"log_file": logPath,
		"log_level": globalLogger.GetLevel().String(),
	}).Info("Logger initialized successfully")
}

// Logger returns the global logger instance
func Logger() *logrus.Logger {
	return globalLogger
}

// WithFields creates a new entry with the specified fields
func WithFields(fields logrus.Fields) *logrus.Entry {
	return globalLogger.WithFields(fields)
}

// WithField creates a new entry with a single field
func WithField(key string, value interface{}) *logrus.Entry {
	return globalLogger.WithField(key, value)
}

// WithError creates a new entry with an error field
func WithError(err error) *logrus.Entry {
	return globalLogger.WithError(err)
}

// Debug logs a debug message
func Debug(args ...interface{}) {
	globalLogger.Debug(args...)
}

// Debugf logs a formatted debug message
func Debugf(format string, args ...interface{}) {
	globalLogger.Debugf(format, args...)
}

// Info logs an info message
func Info(args ...interface{}) {
	globalLogger.Info(args...)
}

// Infof logs a formatted info message
func Infof(format string, args ...interface{}) {
	globalLogger.Infof(format, args...)
}

// Warn logs a warning message
func Warn(args ...interface{}) {
	globalLogger.Warn(args...)
}

// Warnf logs a formatted warning message
func Warnf(format string, args ...interface{}) {
	globalLogger.Warnf(format, args...)
}

// Error logs an error message
func Error(args ...interface{}) {
	globalLogger.Error(args...)
}

// Errorf logs a formatted error message
func Errorf(format string, args ...interface{}) {
	globalLogger.Errorf(format, args...)
}

// Fatal logs a fatal message and exits
func Fatal(args ...interface{}) {
	globalLogger.Fatal(args...)
}

// Fatalf logs a formatted fatal message and exits
func Fatalf(format string, args ...interface{}) {
	globalLogger.Fatalf(format, args...)
}

// Panic logs a panic message and panics
func Panic(args ...interface{}) {
	globalLogger.Panic(args...)
}

// Panicf logs a formatted panic message and panics
func Panicf(format string, args ...interface{}) {
	globalLogger.Panicf(format, args...)
}

// SetLevel sets the log level
func SetLevel(level logrus.Level) {
	globalLogger.SetLevel(level)
}

// GetLevel returns the current log level
func GetLevel() logrus.Level {
	return globalLogger.GetLevel()
}
