package log

import (
	chlog "github.com/charmbracelet/log"
)

const TraceLevel chlog.Level = -5

// LoggerPrefix denotes the current build stage in logs. These prefixes should be used everywhere for logging to keep
// the tool's log output consistent.
type LoggerPrefix string

const (
	LoggerPrefixCompilation LoggerPrefix = "Compilation 🔨"
	LoggerPrefixLocating    LoggerPrefix = "Locating 📡"
	LoggerPrefixLinking     LoggerPrefix = "Linking 🔗"
	LoggerPrefixRendering   LoggerPrefix = "Rendering 🎨"
	LoggerPrefixFormatting  LoggerPrefix = "Formatting 📐"
	LoggerPrefixWriting     LoggerPrefix = "Writing 📝"
)

var loggers = make(map[LoggerPrefix]*Logger)

// GetLogger returns a logger for the given prefix. If the logger does not exist, it is created.
func GetLogger(prefix LoggerPrefix) *Logger {
	if _, ok := loggers[prefix]; !ok {
		loggers[prefix] = &Logger{Logger: chlog.Default().WithPrefix(string(prefix))}
	}
	return loggers[prefix]
}

type Logger struct {
	*chlog.Logger
}

func (l *Logger) Trace(msg interface{}, args ...interface{}) {
	if l.GetLevel() > TraceLevel {
		return
	}
	l.Logger.Debug(msg, args...)
}
