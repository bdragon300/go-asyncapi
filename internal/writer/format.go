package writer

import (
	"bytes"
	"go/format"
	"slices"
	"strings"

	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/samber/lo"
)

// FormatFiles formats the file buffers in-place applying go fmt.
func FormatFiles(files map[string]*bytes.Buffer) error {
	logger := log.GetLogger(log.LoggerPrefixFormatting)

	keys := lo.Keys(files)
	slices.Sort(keys)
	for _, fileName := range keys {
		if !strings.HasSuffix(fileName, ".go") {
			logger.Debug("Skip a file", "name", fileName)
			continue
		}
		buf := files[fileName]
		logger.Debug("File", "name", fileName, "bytes", buf.Len())
		formatted, err := format.Source(buf.Bytes())
		if err != nil {
			return types.MultilineError{err, buf.Bytes()}
		}
		buf.Reset()
		buf.Write(formatted)
		logger.Debug("-> File formatted", "name", fileName, "bytes", buf.Len())
	}

	logger.Info("Formatting complete", "files", len(files))
	return nil
}
