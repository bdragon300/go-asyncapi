package log

import (
	"fmt"
	"io"
	"os"

	chlog "github.com/charmbracelet/log"
)

const TraceLevel chlog.Level = -5

// LoggerPrefix denotes the current build stage in logs. These prefixes should be used everywhere for logging to keep
// the tool's log output consistent.
type LoggerPrefix string

var (
	quietMode bool
	logLevel  chlog.Level
)

func SetQuietMode(quiet bool) {
	quietMode = quiet
}

func SetVerbose(v int) {
	switch v {
	case 0:
		logLevel = chlog.InfoLevel
	case 1:
		logLevel = chlog.DebugLevel
	case 2:
		logLevel = TraceLevel
	default:
		panic("Invalid verbosity level, use 0, 1 or 2")
	}
}

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
		var out io.Writer = os.Stderr
		if quietMode {
			out = io.Discard
		}
		opts := chlog.Options{Prefix: string(prefix), Level: logLevel}
		loggers[prefix] = &Logger{Logger: chlog.NewWithOptions(out, opts), quietMode: quietMode}
	}
	return loggers[prefix]
}

type Logger struct {
	*chlog.Logger
	quietMode bool
}

func (l *Logger) Trace(msg interface{}, args ...interface{}) {
	if l.GetLevel() > TraceLevel {
		return
	}
	l.Debug(msg, args...)
}

func (l *Logger) Error(msg any, args ...any) {
	if l.quietMode {
		// Errors should always be printed to stderr, even in quiet mode.
		_, _ = fmt.Fprintf(os.Stderr, "ERROR: %v\n", msg)
		return
	}
	l.Logger.Error(msg, args...)
}
