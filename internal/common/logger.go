package common

import "github.com/charmbracelet/log"

const TraceLevel log.Level = -5

func NewLogger(prefix string) *Logger {
	return &Logger{Logger: log.Default().WithPrefix(prefix)}
}

type Logger struct {
	*log.Logger
}

func (l *Logger) Trace(msg interface{}, args ...interface{}) {
	if l.GetLevel() > TraceLevel {
		return
	}
	l.Logger.Debug(msg, args...)
}
