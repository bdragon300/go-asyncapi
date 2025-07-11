package writer

import (
	"bytes"
	"fmt"
	"os"
	"path"

	"github.com/bdragon300/go-asyncapi/internal/log"
)

// WriteBuffersToFiles receives the buffers by file name and writes the buffers to files in the baseDir directory.
// Files are truncated if they already exist or created if they don't.
func WriteBuffersToFiles(files map[string]*bytes.Buffer, baseDir string) error {
	logger := log.GetLogger(log.LoggerPrefixWriting)

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

	logger.Info("Writing complete", "files", len(files))
	return nil
}

// ensureDir ensures that the directory at the given path exists. If not, creates it recursively.
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
