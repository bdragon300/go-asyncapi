package writer

import (
	"bytes"
	"fmt"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/samber/lo"
	"go/format"
	"os"
	"path"
	"slices"
)

// FormatFiles formats the files in-place in the map using gofmt
func FormatFiles(files map[string]*bytes.Buffer) error {
	logger := log.GetLogger(log.LoggerPrefixFormatting)
	logger.Info("Run formatting")

	keys := lo.Keys(files)
	slices.Sort(keys)
	for _, fileName := range keys {
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

	logger.Info("Formatting completed", "files", len(files))
	return nil
}

func WriteToFiles(files map[string]*bytes.Buffer, baseDir string) error {
	logger := log.GetLogger(log.LoggerPrefixWriting)
	logger.Info("Run writing")

	if err := ensureDir(baseDir); err != nil {
		return err
	}
	totalBytes := 0
	for fileName, buf := range files {
		logger.Debug("File", "name", fileName)
		fullPath := path.Join(baseDir, fileName)
		if err := ensureDir(path.Dir(fullPath)); err != nil {
			return err
		}

		if err := os.WriteFile(fullPath, buf.Bytes(), 0o644); err != nil {
			return err
		}
		logger.Debug("-> File wrote", "name", fullPath, "bytes", buf.Len())
		totalBytes += buf.Len()
	}
	logger.Debugf("Writer stats: files: %d, total bytes: %d", len(files), totalBytes)

	logger.Info("Writing completed", "files", len(files))
	return nil
}

func ensureDir(path string) error {
	if info, err := os.Stat(path); os.IsNotExist(err) {
		if err2 := os.MkdirAll(path, 0o755); err2 != nil {
			return err2
		}
	} else if err != nil {
		return err
	} else if !info.IsDir() {
		return fmt.Errorf("path %q is not a directory", path)
	}

	return nil
}
