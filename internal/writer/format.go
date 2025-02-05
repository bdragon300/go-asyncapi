package writer

import (
	"bytes"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/samber/lo"
	"go/format"
	"slices"
	"strings"
)

// FormatFiles formats the files in-place in the map using gofmt
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
			return types.ErrorWithContent{err, buf.Bytes()}
		}
		buf.Reset()
		buf.Write(formatted)
		logger.Debug("-> File formatted", "name", fileName, "bytes", buf.Len())
	}

	logger.Info("Formatting complete", "files", len(files))
	return nil
}