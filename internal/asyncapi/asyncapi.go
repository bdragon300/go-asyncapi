// Package asyncapi contains the structs reflecting the AsyncAPI document structure. Every struct has its own
// compilation logic. The compilation logic is recursively invoked by the compiler and produces the compilation
// artifacts -- the types from [render] package.
//
// The output of the compilation stage is:
//
//   - artifacts -- types from [render] package representing the document entities
//   - promises -- placeholders for the compiled artifacts we want to refer to, but that are not ready yet (see below)
//
// All mentioned output is collected in the [common.CompilationStorage] inside the compiler context.
//
// # Entity relations
//
// Every entity compiles *independently* from others, ignoring other entities. For example,
// during compiling the jsonschema, we shouldn't read its fields directly to build an artifact, because they're also a
// jsonschemas, therefore they're other entities.
// Or we can't read a server's properties to build a channel object that refers to this server.
//
// The reason is that the entity we want to refer to may be the $ref instead, that refers to another place in the document
// or even another document somewhere. So to get this object we may need to download the external document, which is
// impossible on this step.
//
// Another reason is that we can't find the entity's compiled artifact and relate to it, because the compiler may invoke the
// compilation logic in any order.
//
// In fact, we need to relate artifacts to each other to generate the right Go code.
//
// To handle this, we use the late binding technique. Simply put, we use [lang.Promise]-like object that acts as an artifact
// placeholder, specify how to find an artifact ($ref or callback) and defer artifact usage to the rendering stage.
// After the compilation stage finishes, the linker resolves promises and assigns artifacts to them.
// So, when the rendering stage starts, we can use the artifacts.
//
// For more information, see [compiler] and [linker] package docs.
package asyncapi

import (
	"fmt"

	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"

	"github.com/bdragon300/go-asyncapi/internal/render"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"golang.org/x/mod/semver"
)

const minAsyncAPIVersion = "3.0.0"

// AsyncAPI is the root document object of the AsyncAPI document. [AsyncAPI specification].
//
// [AsyncAPI specification]: https://github.com/asyncapi/spec/blob/master/spec/asyncapi.md
type AsyncAPI struct {
	Asyncapi           string                              `json:"asyncapi" yaml:"asyncapi"`
	ID                 string                              `json:"id" yaml:"id"`
	Info               InfoItem                            `json:"info" yaml:"info"`
	Servers            types.OrderedMap[string, Server]    `json:"servers" yaml:"servers" cgen:"selectable"`
	DefaultContentType string                              `json:"defaultContentType" yaml:"defaultContentType"`
	Channels           types.OrderedMap[string, Channel]   `json:"channels" yaml:"channels" cgen:"selectable"`
	Operations         types.OrderedMap[string, Operation] `json:"operations" yaml:"operations" cgen:"selectable"`
	Components         ComponentsItem                      `json:"components" yaml:"components"`
}

func (a AsyncAPI) Compile(ctx *compile.Context) error {
	obj, err := a.build(ctx)
	if err != nil {
		return err
	}
	ctx.PutArtifact(obj)
	return nil
}

func (a AsyncAPI) build(ctx *compile.Context) (*render.AsyncAPI, error) {
	ctx.Logger.Trace("AsyncAPI root object")
	res := &render.AsyncAPI{
		DefaultContentType: a.DefaultContentType,
	}

	if a.Asyncapi == "" || !semver.IsValid("v"+a.Asyncapi) {
		return nil, types.CompileError{Err: fmt.Errorf("bad asyncapi version: %q", a.Asyncapi), Path: ctx.CurrentPositionRef()}
	}
	if semver.Compare("v"+a.Asyncapi, "v"+minAsyncAPIVersion) < 0 {
		ctx.Logger.Warn("AsyncAPI version is not supported by the go-asyncapi, the result may contain errors", "version", a.Asyncapi, "minVersion", minAsyncAPIVersion)
	}

	return res, nil
}

type InfoItem struct {
	Title          string                `json:"title" yaml:"title"`
	Version        string                `json:"version" yaml:"version"`
	Description    string                `json:"description" yaml:"description"`
	TermsOfService string                `json:"termsOfService" yaml:"termsOfService"`
	Contact        ContactItem           `json:"contact" yaml:"contact"`
	License        LicenseItem           `json:"license" yaml:"license"`
	Tags           []Tag                 `json:"tags" yaml:"tags"`
	ExternalDocs   ExternalDocumentation `json:"externalDocs" yaml:"externalDocs"`
}

type ContactItem struct {
	Name  string `json:"name" yaml:"name"`
	URL   string `json:"url" yaml:"url"`
	Email string `json:"email" yaml:"email"`
}

type LicenseItem struct {
	Name string `json:"name" yaml:"name"`
	URL  string `json:"url" yaml:"url"`
}

type ExternalDocumentation struct {
	Description string `json:"description" yaml:"description"`
	URL         string `json:"url" yaml:"url"`
}
