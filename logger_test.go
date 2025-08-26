package telebot

import (
	"bytes"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

type CustomTestLogger struct {
	logger *log.Logger
	buffer *bytes.Buffer
}

func NewCustomTestLogger() *CustomTestLogger {
	buffer := &bytes.Buffer{}
	return &CustomTestLogger{
		logger: log.New(buffer, "[CUSTOM] ", 0),
		buffer: buffer,
	}
}

func (l *CustomTestLogger) Debug(msg string, args ...any) {
	l.logger.Printf("[DEBUG] "+msg, args...)
}

func (l *CustomTestLogger) Info(msg string, args ...any) {
	l.logger.Printf("[INFO] "+msg, args...)
}

func (l *CustomTestLogger) Warn(msg string, args ...any) {
	l.logger.Printf("[WARN] "+msg, args...)
}

func (l *CustomTestLogger) Error(msg string, args ...any) {
	l.logger.Printf("[ERROR] "+msg, args...)
}

func (l *CustomTestLogger) Fatal(msg string, args ...any) {
	l.logger.Printf("[FATAL] "+msg, args...)
}

func (l *CustomTestLogger) GetOutput() string {
	return l.buffer.String()
}

func (l *CustomTestLogger) LogMode() LogLevel {
	return LogLevelDebug
}

type LevelTestLogger struct {
	*DefaultLogger
	buffer *bytes.Buffer
}

func NewLevelTestLogger(level LogLevel) *LevelTestLogger {
	buffer := &bytes.Buffer{}
	defaultLogger := &DefaultLogger{
		logger:  log.New(buffer, "[LEVEL] ", 0),
		enabled: true,
		level:   level,
	}
	return &LevelTestLogger{
		DefaultLogger: defaultLogger,
		buffer:        buffer,
	}
}

func (l *LevelTestLogger) GetOutput() string {
	return l.buffer.String()
}

func TestDefaultLogger(t *testing.T) {
	pref := Settings{
		Offline: true,
		Log: &LogConfig{
			Enable: true,
		},
	}

	bot, err := NewBot(pref)
	assert.NoError(t, err)
	assert.NotNil(t, bot.logger)

	update := Update{
		Message: &Message{
			Text:   "/start",
			Sender: &User{ID: 123, FirstName: "Test"},
		},
	}
	ctx := NewContext(bot, update)

	logger := ctx.Logger()
	assert.NotNil(t, logger)

	assert.NotPanics(t, func() {
		logger.Debug("Debug message")
		logger.Info("Info message")
		logger.Warn("Warn message")
		logger.Error("Error message")
	})
}

func TestNoOpLogger(t *testing.T) {
	pref := Settings{
		Offline: true,
		Log: &LogConfig{
			Enable: false,
		},
	}

	bot, err := NewBot(pref)
	assert.NoError(t, err)
	assert.NotNil(t, bot.logger)

	update := Update{
		Message: &Message{
			Text:   "/silent",
			Sender: &User{ID: 456, FirstName: "Silent"},
		},
	}
	ctx := NewContext(bot, update)

	logger := ctx.Logger()
	assert.NotNil(t, logger)

	assert.NotPanics(t, func() {
		logger.Debug("This won't be logged")
		logger.Info("Neither will this")
		logger.Warn("Nor this")
		logger.Error("Or this")
	})
}

func TestCustomLogger(t *testing.T) {
	customLogger := NewCustomTestLogger()
	pref := Settings{
		Offline: true,
		Log: &LogConfig{
			Enable: true,
			Logger: customLogger,
		},
	}

	bot, err := NewBot(pref)
	assert.NoError(t, err)
	assert.Equal(t, customLogger, bot.logger)

	update := Update{
		Message: &Message{
			Text:   "/custom",
			Sender: &User{ID: 789, FirstName: "Custom"},
		},
	}
	ctx := NewContext(bot, update)

	logger := ctx.Logger()
	assert.Equal(t, customLogger, logger)

	logger.Info("Custom logger test message")
	output := customLogger.GetOutput()
	assert.Contains(t, output, "[CUSTOM]")
	assert.Contains(t, output, "[INFO]")
	assert.Contains(t, output, "Custom logger test message")
}

func TestLoggerInHandler(t *testing.T) {
	customLogger := NewCustomTestLogger()
	pref := Settings{
		Offline: true,
		Log: &LogConfig{
			Enable: true,
			Logger: customLogger,
		},
	}

	bot, err := NewBot(pref)
	assert.NoError(t, err)

	tp := newTestPoller()

	bot.Handle("/test", func(c Context) error {
		c.Logger().Info("Handler received message from user %d", c.Sender().ID)
		c.Logger().Debug("Message text: %s", c.Text())
		tp.done <- struct{}{}
		return nil
	})

	update := Update{
		Message: &Message{
			Text:   "/test",
			Sender: &User{ID: 12345, FirstName: "TestUser"},
		},
	}

	go bot.Start()

	bot.ProcessUpdate(update)

	<-tp.done
	bot.Stop()

	output := customLogger.GetOutput()
	assert.Contains(t, output, "[INFO]")
	assert.Contains(t, output, "Handler received message from user 12345")
	assert.Contains(t, output, "[DEBUG]")
	assert.Contains(t, output, "Message text: /test")
}

func TestLogConfigWithDisabledLogging(t *testing.T) {
	customLogger := NewCustomTestLogger()
	pref := Settings{
		Offline: true,
		Log: &LogConfig{
			Enable: false,
			Logger: customLogger,
		},
	}

	bot, err := NewBot(pref)
	assert.NoError(t, err)

	assert.IsType(t, &NoOpLogger{}, bot.logger)
	assert.NotEqual(t, customLogger, bot.logger)
}

func TestLogConfigWithLogLevel(t *testing.T) {
	levelLogger := NewLevelTestLogger(LogLevelWarn)
	pref := Settings{
		Offline: true,
		Log: &LogConfig{
			Enable: true,
			Logger: levelLogger,
		},
	}

	bot, err := NewBot(pref)
	assert.NoError(t, err)
	assert.NotNil(t, bot.logger)
	assert.Equal(t, levelLogger, bot.logger)

	update := Update{
		Message: &Message{
			Text:   "/level",
			Sender: &User{ID: 999, FirstName: "LevelTest"},
		},
	}
	ctx := NewContext(bot, update)

	logger := ctx.Logger()
	assert.NotNil(t, logger)
	assert.Equal(t, LogLevelWarn, logger.LogMode())

	logger.Debug("This should not be logged")
	logger.Info("This should not be logged")
	logger.Warn("This should be logged")
	logger.Error("This should be logged")

	output := levelLogger.GetOutput()
	assert.NotContains(t, output, "This should not be logged")
	assert.Contains(t, output, "This should be logged")
	assert.Contains(t, output, "[WARN]")
	assert.Contains(t, output, "[ERROR]")
}

func TestLogConfigNilUsesNoOpLogger(t *testing.T) {
	pref := Settings{
		Offline: true,
		Log:     nil,
	}

	bot, err := NewBot(pref)
	assert.NoError(t, err)
	assert.IsType(t, &NoOpLogger{}, bot.logger)
}
