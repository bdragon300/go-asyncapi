package packages

import (
	"github.com/bdragon300/asyncapi-codegen/assets"
	"github.com/bdragon300/asyncapi-codegen/internal/common"
	"github.com/bdragon300/asyncapi-codegen/internal/utils"
)

type RuntimePackage struct{}

func (r RuntimePackage) Put(_ *common.CompileContext, _ common.Assembler) {
	panic("implement me")
}

func (r RuntimePackage) Items() []common.PackageItem[common.Assembler] {
	return nil
}

func (r RuntimePackage) FindBy(_ func(item any, path []string) bool) (common.Assembler, bool) {
	return nil, false
}

func (r RuntimePackage) ListBy(_ func(item any, path []string) bool) []common.Assembler {
	return nil
}

func RenderRuntime(_ *RuntimePackage, baseDir string) error {
	return utils.CopyRecursive(assets.AssetsFS, baseDir)
}
