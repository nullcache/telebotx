package telebot

import (
	"log"
	"os"
)

// LogLevel represents the logging level
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
	LogLevelOff
)

// String returns the string representation of the log level
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "DEBUG"
	case LogLevelInfo:
		return "INFO"
	case LogLevelWarn:
		return "WARN"
	case LogLevelError:
		return "ERROR"
	case LogLevelFatal:
		return "FATAL"
	case LogLevelOff:
		return "OFF"
	default:
		return "UNKNOWN"
	}
}

// LogConfig represents the logging configuration
type LogConfig struct {
	// Enable controls whether logging is enabled
	// This has the highest priority - if false, no logging will occur regardless of other settings
	Enable bool

	// Level controls the minimum log level to output
	Level LogLevel

	// Prefix is the prefix for log messages
	Prefix string

	// Logger is the logger implementation to use.
	Logger Logger
}

// Logger represents a generic logging interface that can be implemented
// by different logging libraries or custom implementations.
type Logger interface {
	Debug(msg string, args ...any)
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Error(msg string, args ...any)
	Fatal(msg string, args ...any)
	LogMode() LogLevel
}

type DefaultLogger struct {
	logger  *log.Logger
	enabled bool
	level   LogLevel
}

// NewDefaultLogger creates a new DefaultLogger instance with custom configuration.
func NewDefaultLogger(level LogLevel, prefix string) *DefaultLogger {
	return &DefaultLogger{
		logger:  log.New(os.Stdout, prefix, log.LstdFlags|log.Lshortfile),
		enabled: true,
		level:   level,
	}
}

// Debug logs a debug message
func (l *DefaultLogger) Debug(msg string, args ...any) {
	if !l.enabled || l.level > LogLevelDebug {
		return
	}
	l.logger.Printf("[DEBUG] "+msg, args...)
}

// Info logs an info message
func (l *DefaultLogger) Info(msg string, args ...any) {
	if !l.enabled || l.level > LogLevelInfo {
		return
	}
	l.logger.Printf("[INFO] "+msg, args...)
}

// Warn logs a warning message
func (l *DefaultLogger) Warn(msg string, args ...any) {
	if !l.enabled || l.level > LogLevelWarn {
		return
	}
	l.logger.Printf("[WARN] "+msg, args...)
}

// Error logs an error message
func (l *DefaultLogger) Error(msg string, args ...any) {
	if !l.enabled || l.level > LogLevelError {
		return
	}
	l.logger.Printf("[ERROR] "+msg, args...)
}

// Fatal logs a fatal message and exits
func (l *DefaultLogger) Fatal(msg string, args ...any) {
	if !l.enabled || l.level > LogLevelFatal {
		return
	}
	l.logger.Printf("[FATAL] "+msg, args...)
	os.Exit(1)
}

// LogMode returns the current log level
func (l *DefaultLogger) LogMode() LogLevel {
	if !l.enabled {
		return LogLevelOff
	}
	return l.level
}

// NoOpLogger is a logger that does nothing. Useful when logging is disabled.
type NoOpLogger struct{}

// NewNoOpLogger creates a new NoOpLogger instance
func NewNoOpLogger() *NoOpLogger {
	return &NoOpLogger{}
}

// Debug does nothing
func (l *NoOpLogger) Debug(msg string, args ...any) {}

// Info does nothing
func (l *NoOpLogger) Info(msg string, args ...any) {}

// Warn does nothing
func (l *NoOpLogger) Warn(msg string, args ...any) {}

// Error does nothing
func (l *NoOpLogger) Error(msg string, args ...any) {}

// Fatal does nothing
func (l *NoOpLogger) Fatal(msg string, args ...any) {}

// LogMode returns LogLevelOff since this logger does nothing
func (l *NoOpLogger) LogMode() LogLevel {
	return LogLevelOff
}

// StdLogger wraps Go's standard log.Logger to implement our Logger interface
type StdLogger struct {
	logger  *log.Logger
	enabled bool
}

// NewStdLogger creates a new StdLogger that wraps the provided log.Logger
func NewStdLogger(logger *log.Logger, enabled bool) *StdLogger {
	if logger == nil {
		logger = log.Default()
	}
	return &StdLogger{
		logger:  logger,
		enabled: enabled,
	}
}

// Debug logs a debug message
func (l *StdLogger) Debug(msg string, args ...any) {
	if !l.enabled {
		return
	}
	l.logger.Printf("[DEBUG] "+msg, args...)
}

// Info logs an info message
func (l *StdLogger) Info(msg string, args ...any) {
	if !l.enabled {
		return
	}
	l.logger.Printf("[INFO] "+msg, args...)
}

// Warn logs a warning message
func (l *StdLogger) Warn(msg string, args ...any) {
	if !l.enabled {
		return
	}
	l.logger.Printf("[WARN] "+msg, args...)
}

// Error logs an error message
func (l *StdLogger) Error(msg string, args ...any) {
	if !l.enabled {
		return
	}
	l.logger.Printf("[ERROR] "+msg, args...)
}

// Fatal logs a fatal message and exits
func (l *StdLogger) Fatal(msg string, args ...any) {
	if !l.enabled {
		return
	}
	l.logger.Fatalf("[FATAL] "+msg, args...)
}

// LogMode returns the current log level (StdLogger doesn't support level filtering, so always return Debug when enabled)
func (l *StdLogger) LogMode() LogLevel {
	if !l.enabled {
		return LogLevelOff
	}
	return LogLevelDebug
}

// NewLogger creates a logger based on the provided LogConfig
func NewLogger(config LogConfig) Logger {
	// Enable has the highest priority
	if !config.Enable {
		return NewNoOpLogger()
	}

	// If a custom logger is provided, use it
	if config.Logger != nil {
		return config.Logger
	}

	// Create default logger with configuration
	return NewDefaultLogger(config.Level, config.Prefix)
}
