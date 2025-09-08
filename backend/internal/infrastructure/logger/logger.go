package logger

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/terminator791/t-pos/internal/infrastructure/config"
)

// Logger interface
type Logger interface {
	Info(msg string, args ...interface{})
	Error(msg string, args ...interface{})
	Debug(msg string, args ...interface{})
	Warn(msg string, args ...interface{})
}

// SimpleLogger implements Logger interface
type SimpleLogger struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	debugLogger *log.Logger
	warnLogger  *log.Logger
}

// NewLogger creates a new logger instance
func NewLogger(cfg *config.Config) (Logger, error) {
	// Create logs directory if it doesn't exist
	logDir := filepath.Dir(cfg.Logger.Filename)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, err
	}

	// Open log file
	logFile, err := os.OpenFile(cfg.Logger.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	// Create multi-writer to write to both file and stdout
	multiWriter := io.MultiWriter(os.Stdout, logFile)

	return &SimpleLogger{
		infoLogger:  log.New(multiWriter, "INFO: ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger: log.New(multiWriter, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile),
		debugLogger: log.New(multiWriter, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile),
		warnLogger:  log.New(multiWriter, "WARN: ", log.Ldate|log.Ltime|log.Lshortfile),
	}, nil
}

// Info logs info messages
func (l *SimpleLogger) Info(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.infoLogger.Printf(msg, args...)
	} else {
		l.infoLogger.Println(msg)
	}
}

// Error logs error messages
func (l *SimpleLogger) Error(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.errorLogger.Printf(msg, args...)
	} else {
		l.errorLogger.Println(msg)
	}
}

// Debug logs debug messages
func (l *SimpleLogger) Debug(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.debugLogger.Printf(msg, args...)
	} else {
		l.debugLogger.Println(msg)
	}
}

// Warn logs warning messages
func (l *SimpleLogger) Warn(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.warnLogger.Printf(msg, args...)
	} else {
		l.warnLogger.Println(msg)
	}
}