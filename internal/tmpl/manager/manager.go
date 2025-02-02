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
	return &TemplateRenderManager{
		RenderOpts: opts,
		stateCommitted: make(map[string]FileRenderState),
		Buffer: new(bytes.Buffer),
		ImportsManager: new(ImportsManager),
		NamespaceManager: new(NamespaceManager),
		namespaceCommitted: new(NamespaceManager),
	}
}

type TemplateRenderManager struct {
	RenderOpts common.RenderOpts

	FileName string
	PackageName     string
	Buffer           *bytes.Buffer

	CurrentObject   common.Renderable
	CurrentSelection common.ConfigSelectionItem

	ImportsManager  *ImportsManager
	NamespaceManager *NamespaceManager
	Implementations  []ImplementationItem

	stateCommitted           map[string]FileRenderState
	namespaceCommitted       *NamespaceManager
	implementationsCommitted []ImplementationItem
}

func (r *TemplateRenderManager) BeginFile(fileName, packageName string) {
	if _, ok := r.stateCommitted[fileName]; !ok {
		pkgName, _ := lo.Coalesce(packageName, utils.GetPackageName(path.Dir(fileName)))
		r.stateCommitted[fileName] = FileRenderState{
			PackageName: pkgName,
			FileName: fileName,
			Buffer: new(bytes.Buffer),
			Imports: new(ImportsManager),
		}
	}
	state := r.stateCommitted[fileName]

	r.FileName = state.FileName
	r.ImportsManager = state.Imports.Clone()
	r.PackageName = state.PackageName
	r.NamespaceManager = r.namespaceCommitted.Clone()
	r.Implementations = slices.Clone(r.implementationsCommitted)

	r.Buffer.Reset()
}

func (r *TemplateRenderManager) BeginCodeObject(obj common.Renderable, selection common.ConfigSelectionItem) {
	r.CurrentObject = obj
	r.CurrentSelection = selection
}

func (r *TemplateRenderManager) AddImplementation(obj common.ImplementationObject, directory string) {
	r.Implementations = append(r.Implementations, ImplementationItem{Object: obj, Directory: directory})
}

func (r *TemplateRenderManager) Commit() {
	r.namespaceCommitted = r.NamespaceManager
	r.implementationsCommitted = r.Implementations

	if r.FileName != "" {
		state := r.stateCommitted[r.FileName]
		state.Imports = r.ImportsManager
		lo.Must(state.Buffer.ReadFrom(r.Buffer))
		state.Buffer.WriteRune('\n') // Separate writes following each other with newline
		r.stateCommitted[r.FileName] = state
	}

	r.CurrentObject = nil
	r.FileName = ""
}

func (r *TemplateRenderManager) CommittedStates() map[string]FileRenderState {
	return r.stateCommitted
}

func (r *TemplateRenderManager) CommittedImplementations() []ImplementationItem {
	return r.implementationsCommitted
}

type ImplementationItem struct {
	Object    common.ImplementationObject
	Directory string // Evaluated template expression
}

type FileRenderState struct {
	Imports     *ImportsManager
	PackageName string
	FileName string
	Buffer   *bytes.Buffer
}