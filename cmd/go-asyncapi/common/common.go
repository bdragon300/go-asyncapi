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
	ErrWrongCliArgs      = errors.New("cli args")
	ErrInterruptedByUser = fmt.Errorf("interrupted by user")
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
