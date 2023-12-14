package factory

import (
	"fmt"
	"net/url"

	"github.com/bdragon300/asyncapi-codegen-go/internal/common"
	"github.com/bdragon300/asyncapi-codegen-go/internal/compiler"
)

type Compiler interface {
	Protocols() []string
	DirectRenderItems(pkgName string) []compiler.PackageItem
	Packages() []string
	AllItems() []compiler.PackageItem
	Load() error
	Compile(linker common.Linker) error
	Stats() string
}

func MakeCompiler(filePath string) (Compiler, error) {
	u, err := url.Parse(filePath)
	if err != nil {
		return nil, fmt.Errorf("parsing the path: %w", err)
	}

	if u.Scheme == "file" || u.Host == "" && u.User == nil && u.Scheme == "" {
		// Cut out the optional scheme, assume that the rest is filename (path can include '#' for example)
		u.Scheme = ""
		return compiler.NewLocalFile(u.String()), nil
	}
	panic("remote compiler not implemented")
}
