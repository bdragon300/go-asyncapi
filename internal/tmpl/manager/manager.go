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

// TemplateRenderManager provides the transactional model for the rendering the files. It manages the states
// of every file being rendered. It also used by template functions.
//
// File state includes the file name, package name, contents buffer, imports, and template namespace. Manager starts
// exposing the state of a particular file by BeginFile call and keeps the committed states of all files, where
// changes in the exposed state are saved on Commit call.
//
// The typical workflow is:
//
//  1. Initialize manager with rendering options
//  2. Load the templates, assign the TemplateLoader
//  3. Call BeginFile for every file being rendered. Manager set the state of that file as current (creating a new state
//     for a new file).
//  4. Invoke the renderer code (which invokes the template code, template functions and so on), that makes changes
//     to the current state.
//  5. Call Commit to save the changes to the current file state on success. Otherwise, the changes are discarded.
//  6. Repeat steps 3-5 any number of times.
//  7. Gather the committed results.
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

// BeginFile sets the current state to be exposed by manager for a file. If the file was not rendered before,
// it creates a new state with given package name. All uncommitted changes are discarded.
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

// AddImplementation adds a new implementation to the list. Gets to the committed list on Commit call, otherwise
// the changes are discarded.
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
