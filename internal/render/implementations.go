package render

import (
	"io/fs"

	"github.com/bdragon300/asyncapi-codegen/implementations"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
)

func WriteImplementation(implDir, baseDir string) error {
	subDir, err := fs.Sub(implementations.Implementations, implDir)
	if err != nil {
		return err
	}

	return utils.CopyRecursive(subDir, baseDir)
}
