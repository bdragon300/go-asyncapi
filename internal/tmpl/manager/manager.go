package manager

import (
	"bytes"
	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
	"path"
	"slices"
)

func NewTemplateRenderManager(opts common.RenderOpts) *TemplateRenderManager {
	return &TemplateRenderManager{RenderOpts: opts, stateCommitted: make(map[string]FileRenderState)}
}

type TemplateRenderManager struct {
	RenderOpts common.RenderOpts

	FileName string
	PackageName     string
	Buffer           bytes.Buffer

	ImportsManager  ImportsManager
	NamespaceManager NamespaceManager
	Implementations  []ImplementationItem

	CurrentObject   common.Renderable
	CurrentSelection common.ConfigSelectionItem

	stateCommitted           map[string]FileRenderState
	namespaceCommitted       NamespaceManager
	implementationsCommitted []ImplementationItem
}

func (r *TemplateRenderManager) BeginObject(obj common.Renderable, selection common.ConfigSelectionItem, fileName string) {
	if _, ok := r.stateCommitted[fileName]; !ok {
		pkgName, _ := lo.Coalesce(selection.Render.Package, utils.GetPackageName(path.Dir(fileName)))
		r.stateCommitted[fileName] = FileRenderState{PackageName: pkgName, FileName: fileName}
	}
	state := r.stateCommitted[fileName]

	r.ImportsManager = state.Imports.Clone()
	r.PackageName = state.PackageName
	r.FileName = state.FileName
	r.NamespaceManager = r.namespaceCommitted.Clone()
	r.Implementations = slices.Clone(r.implementationsCommitted)

	r.Buffer.Reset()

	r.CurrentObject = obj
	r.CurrentSelection = selection
}

func (r *TemplateRenderManager) Commit() {
	r.namespaceCommitted = r.NamespaceManager
	r.implementationsCommitted = r.Implementations

	state := r.stateCommitted[r.FileName]
	state.Imports = r.ImportsManager
	lo.Must(state.Buffer.ReadFrom(&r.Buffer))
	state.Buffer.WriteRune('\n') // Separate writes following each other with newline
	r.stateCommitted[r.FileName] = state

	r.CurrentObject = nil
}

func (r *TemplateRenderManager) AddImplementation(obj common.ImplementationObject, directory string) {
	r.Implementations = append(r.Implementations, ImplementationItem{Object: obj, Directory: directory})
}

func (r *TemplateRenderManager) AllStates() map[string]FileRenderState {
	return r.stateCommitted
}

type ImplementationItem struct {
	Object    common.ImplementationObject
	Directory string // Evaluated template expression
}

type FileRenderState struct {
	Imports     ImportsManager
	PackageName string
	FileName string
	Buffer   bytes.Buffer
}