package manager

import (
	"bytes"
	"path"
	"slices"
	"text/template"

	"github.com/bdragon300/go-asyncapi/internal/common"
	"github.com/bdragon300/go-asyncapi/internal/utils"
	"github.com/samber/lo"
)

// NewTemplateRenderManager returns a new instance of TemplateRenderManager with given rendering options.
func NewTemplateRenderManager(opts common.RenderOpts) *TemplateRenderManager {
	return &TemplateRenderManager{
		RenderOpts:         opts,
		Buffer:             new(bytes.Buffer),
		ImportsManager:     new(ImportsManager),
		NamespaceManager:   new(NamespaceManager),
		stateCommitted:     make(map[string]FileRenderState),
		namespaceCommitted: new(NamespaceManager),
	}
}

type templateLoader interface {
	LoadTemplate(name string) (*template.Template, error)
	LoadRootTemplate() (*template.Template, error)
}

// TemplateRenderManager provides the transactional model for the rendering the files. It provides a way to
// make changes in the files with the ability to roll back to the last committed state.
//
// File state includes the file name, package name, contents buffer, imports, and template namespace.
//
// The typical workflow is load-write-commit cycle:
//
//  1. Call BeginFile to load the committed state (or creates a new one) of given file. After that the file state
//     starts to be exposed using the manager methods and fields.
//  2. Write the content to file buffer or make other changes in the its state. These changes are considered as not
//     committed yet and will be rolled back on next BeginFile call.
//  3. Call Commit to commit file changes.
//
// After all files are rendered, the committed states can be retrieved using Committed* methods.
//
// TODO: Refactor, split in smaller parts. Separate the transactions as generic and separate implementations, current state expose using methods.
type TemplateRenderManager struct {
	// RenderOpts keeps the rendering options
	RenderOpts common.RenderOpts

	// CurrentObject is an object being currently rendered
	CurrentObject common.Artifact
	// CurrentLayoutItem is a layout item that was used to select the CurrentObject
	CurrentLayoutItem common.ConfigLayoutItem
	TemplateLoader    templateLoader

	// File state. The following fields are restored from committed
	FileName    string
	PackageName string
	// Buffer is write-only file contents buffer. When Commit is called, it appends to the committed file contents.
	Buffer *bytes.Buffer
	// ImportsManager keeps the imports list for the current file
	ImportsManager *ImportsManager
	// NamespaceManager keeps the template definitions namespace for the current file
	NamespaceManager *NamespaceManager
	// Implementations keeps the current list of implementations
	Implementations []ImplementationItem

	// Committed state
	stateCommitted           map[string]FileRenderState
	namespaceCommitted       *NamespaceManager
	implementationsCommitted []ImplementationItem
}

// BeginFile loads the committed state of given file into the manager fields, discarding any uncommitted changes.
// If the file is not loaded yet, it creates a new state.
func (r *TemplateRenderManager) BeginFile(fileName, packageName string) {
	if _, ok := r.stateCommitted[fileName]; !ok {
		pkgName, _ := lo.Coalesce(packageName, utils.GetPackageName(path.Dir(fileName)))
		r.stateCommitted[fileName] = FileRenderState{
			PackageName: pkgName,
			FileName:    fileName,
			Buffer:      new(bytes.Buffer),
			Imports:     new(ImportsManager),
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

// SetCodeObject is helper that just sets the CurrentObject and CurrentLayoutItem fields.
func (r *TemplateRenderManager) SetCodeObject(obj common.Artifact, layoutItem common.ConfigLayoutItem) {
	r.CurrentObject = obj
	r.CurrentLayoutItem = layoutItem
}

// AddImplementation adds a new implementation to the list, making uncommitted changes.
func (r *TemplateRenderManager) AddImplementation(obj common.ImplementationObject, directory string) {
	r.Implementations = append(r.Implementations, ImplementationItem{Object: obj, Directory: directory})
}

// Commit saves the current state to the committed state.
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

// CommittedStates returns the committed states of all files.
func (r *TemplateRenderManager) CommittedStates() map[string]FileRenderState {
	return r.stateCommitted
}

// CommittedImplementations returns the committed implementations list.
func (r *TemplateRenderManager) CommittedImplementations() []ImplementationItem {
	return r.implementationsCommitted
}

type ImplementationItem struct {
	Object    common.ImplementationObject
	Directory string // Evaluated template expression
}

type FileRenderState struct {
	FileName    string
	PackageName string
	Buffer      *bytes.Buffer
	Imports     *ImportsManager
}
