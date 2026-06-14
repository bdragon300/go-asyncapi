package common2

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/locator"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/samber/lo"
)

const (
	DefaultConfigFileName                   = "default_config.yaml"
	DefaultMainTemplateName                 = "main.tmpl"
	DefaultSubprocessLocatorShutdownTimeout = 3 * time.Second
)

var (
	// ErrWrongCliArgs is returned when the CLI arguments are failed to validate. This error is fatal, it causes
	// displaying the help message and exiting with error code.
	ErrWrongCliArgs = errors.New("cli args")

	// ErrInterruptedByUser is returned when the user interrupts the execution. Command exits with code 0 in this case.
	ErrInterruptedByUser = fmt.Errorf("interrupted by user")

	// ErrBadResult indicates that the command has finished without any issues but its result is error.
	// This is not a fatal error, it just causes exiting with error code.
	ErrBadResult = fmt.Errorf("bad result")
)

type DocumentLocator interface {
	Locate(docURL *jsonpointer.JSONPointer) (io.ReadCloser, error)
	ResolveURL(base, target *jsonpointer.JSONPointer) (*jsonpointer.JSONPointer, error)
}

func GetLocator(conf ToolConfig) DocumentLocator {
	logger := log.GetLogger(log.LoggerPrefixLocating)
	if conf.Locator.Command != "" {
		return locator.Subprocess{
			CommandLine:     conf.Locator.Command,
			RunTimeout:      conf.Locator.Timeout,
			ShutdownTimeout: DefaultSubprocessLocatorShutdownTimeout,
			RootDirectory:   conf.Locator.RootDirectory,
			Logger:          logger,
		}
	}
	res := locator.Default{
		Client:        &http.Client{Timeout: conf.Locator.Timeout},
		RootDirectory: conf.Locator.RootDirectory,
		Logger:        logger,
	}
	return res
}

// Coalesce return the first non-zero value from the list of arguments.
func Coalesce[T comparable](vals ...T) T {
	res, _ := lo.Coalesce(vals...)
	return res
}
