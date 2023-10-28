package render

import (
	"fmt"
	"io/fs"
	"os"

	"github.com/bdragon300/asyncapi-codegen-go/implementations"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
)

func WriteImplementation(implDir, baseDir string) error {
	if err := os.MkdirAll(baseDir, 0o750); err != nil {
		return fmt.Errorf("cannot create directory %q: %w", baseDir, err)
	}

	subDir, err := fs.Sub(implementations.Implementations, implDir)
	if err != nil {
		return err
	}

	return utils.CopyRecursive(subDir, baseDir)
}
