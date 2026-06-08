package log

import (
	"encoding/json"
	"fmt"
	"io"
	stdlog "log"
	"os"
	"time"
)

type Level string

const (
	DEBUG Level = "DEBUG"
	INFO  Level = "INFO"
	WARN  Level = "WARN"
	ERROR Level = "ERROR"
)

type Logger struct {
	out   io.Writer
	level Level
}

type entry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
}

var defaultLogger = New(INFO)

func New(level Level) *Logger {
	return &Logger{out: os.Stdout, level: level}
}

func SetDefault(l *Logger) { defaultLogger = l }

func (l *Logger) log(level Level, msg string, extra map[string]interface{}) {
	if levelPriority(level) < levelPriority(l.level) {
		return
	}
	e := entry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     string(level),
		Message:   msg,
		Extra:     extra,
	}
	data, _ := json.Marshal(e)
	fmt.Fprintln(l.out, string(data))
}

func (l *Logger) Debug(msg string, extra map[string]interface{}) { l.log(DEBUG, msg, extra) }
func (l *Logger) Info(msg string, extra map[string]interface{})  { l.log(INFO, msg, extra) }
func (l *Logger) Warn(msg string, extra map[string]interface{})  { l.log(WARN, msg, extra) }
func (l *Logger) Error(msg string, extra map[string]interface{}) { l.log(ERROR, msg, extra) }

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DEBUG, fmt.Sprintf(format, args...), nil)
}
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(INFO, fmt.Sprintf(format, args...), nil)
}
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WARN, fmt.Sprintf(format, args...), nil)
}
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ERROR, fmt.Sprintf(format, args...), nil)
}

func (l *Logger) Fatal(msg string, extra map[string]interface{}) {
	l.log(ERROR, msg, extra)
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log(ERROR, fmt.Sprintf(format, args...), nil)
	os.Exit(1)
}

func levelPriority(l Level) int {
	switch l {
	case DEBUG:
		return 0
	case INFO:
		return 1
	case WARN:
		return 2
	case ERROR:
		return 3
	default:
		return 1
	}
}

// Package-level convenience functions using the default logger.
func Debugf(format string, args ...interface{})  { defaultLogger.Debugf(format, args...) }
func Infof(format string, args ...interface{})   { defaultLogger.Infof(format, args...) }
func Warnf(format string, args ...interface{})   { defaultLogger.Warnf(format, args...) }
func Errorf(format string, args ...interface{})  { defaultLogger.Errorf(format, args...) }
func Debug(msg string, extra map[string]interface{}) { defaultLogger.Debug(msg, extra) }
func Info(msg string, extra map[string]interface{})  { defaultLogger.Info(msg, extra) }
func Warn(msg string, extra map[string]interface{})  { defaultLogger.Warn(msg, extra) }
func Error(msg string, extra map[string]interface{}) { defaultLogger.Error(msg, extra) }
func Fatal(msg string, extra map[string]interface{}) { defaultLogger.Fatal(msg, extra) }
func Fatalf(format string, args ...interface{})      { defaultLogger.Fatalf(format, args...) }

// Default returns a standard library logger compatible wrapper for INFO and above.
func Default() *stdlog.Logger {
	l := New(INFO)
	return stdlog.New(&logWriter{l}, "", 0)
}

type logWriter struct{ l *Logger }

func (w *logWriter) Write(p []byte) (int, error) {
	w.l.Info(string(p), nil)
	return len(p), nil
}
