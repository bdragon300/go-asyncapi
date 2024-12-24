package log

import (
	chlog "github.com/charmbracelet/log"
)

const TraceLevel chlog.Level = -5

type LoggerPrefix string

const (
	LoggerPrefixCompilation LoggerPrefix = "Compilation 🔨"
	LoggerPrefixResolving LoggerPrefix = "Resolving 📡"
	LoggerPrefixLinking LoggerPrefix = "Linking 🔗"
	LoggerPrefixRendering LoggerPrefix = "Rendering 🎨"
	LoggerPrefixFormatting LoggerPrefix = "Formatting 📐"
	LoggerPrefixWriting LoggerPrefix = "Writing 📝"
)

var loggers = make(map[LoggerPrefix]*Logger)

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
