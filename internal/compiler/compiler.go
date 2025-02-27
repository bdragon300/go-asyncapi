// Package compiler contains the logic of the compilation stage of the AsyncAPI document.
//
// The compilation stage consists of two main steps:
//
//  1. Unmarshalling a given AsyncAPI document to the [asyncapi] package structures.
//  2. Recursively walking through these structures and invoking their compilation logic.
//
// Roughly, the compiler does the following:
//
//   - At the start, the compilation queue contains only the root document.
//   - Take a document from the compilation queue, and run the [locator.Default] to get the document
//     data, and unmarshal it to the [asyncapi.AsyncAPI] structure.
//   - Create the [compile.Context] that holds the state of the compilation process and is passed to compilation logic.
//   - Recursively invoke the compilation logic on [asyncapi.AsyncAPI], that the artifacts, promises, filling up the [Document]
//     with them. If an error occurs, the process aborts.
//   - If we meet the $ref to the external document on the way, the compiler adds the path to this document to the compilation
//     queue.
//   - Repeat the process until the compilation queue is empty.
//   - After the compiler invokes all the compilation logic in structures, the compilation stage is finished.
//
// Compiler processes one document at a time, producing one [Document] per document. As a result of this process, we have
// one or several [Document] objects.
// This could slightly remind the simplified version of the "translation unit" concept in the C/C++ compilers.
//
// Next stage, linking, will resolve all promises in all documents, after that, the result can be fed to the rendering
// stage.
//
// See [linker] and [renderer] package docs for more information. See [asyncapi] package description for more information
// about promises.
package compiler

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"reflect"
	"time"

	"github.com/bdragon300/go-asyncapi/internal/compiler/compile"

	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/log"

	"gopkg.in/yaml.v3"

	"github.com/bdragon300/go-asyncapi/internal/common"
)

// NewDocument returns a new compilation document referenced by the given URL.
func NewDocument(u *jsonpointer.JSONPointer) *Document {
	return &Document{
		logger: log.GetLogger(log.LoggerPrefixCompilation),
		url:    u,
	}
}

// Document is a compilation unit, that contains the unmarshalled object tree, metadata. Also, it keeps the
// artifacts and promises that are produced on the compilation stage.
type Document struct {
	url    *jsonpointer.JSONPointer
	logger *log.Logger

	kind        DocumentKind
	objectsTree compiledObject

	// Compilation results
	externalRefs []*jsonpointer.JSONPointer
	artifacts    []common.Artifact
	promises     []common.ObjectPromise
	listPromises []common.ObjectListPromise
}

func (c *Document) AddArtifact(a common.Artifact) {
	c.artifacts = append(c.artifacts, a)
}

func (c *Document) AddExternalRef(ref *jsonpointer.JSONPointer) {
	c.externalRefs = append(c.externalRefs, ref)
}

func (c *Document) AddPromise(p common.ObjectPromise) {
	c.promises = append(c.promises, p)
}

func (c *Document) AddListPromise(p common.ObjectListPromise) {
	c.listPromises = append(c.listPromises, p)
}

func (c *Document) DocumentURL() jsonpointer.JSONPointer {
	return *c.url
}

func (c *Document) Artifacts() []common.Artifact {
	return c.artifacts
}

func (c *Document) Promises() []common.ObjectPromise {
	return c.promises
}

func (c *Document) ListPromises() []common.ObjectListPromise {
	return c.listPromises
}

type documentLocator interface {
	Locate(documentURL *jsonpointer.JSONPointer) (io.ReadCloser, error)
}

// Load loads the document using the given locator, reading the unmarshalled contents and metadata
func (c *Document) Load(locator documentLocator) error {
	c.logger.Debug("Locate and load a document", "url", c.url)
	buf, err := c.readFile(c.url, locator)
	if err != nil {
		return fmt.Errorf("read file: %w", err)
	}

	c.logger.Trace("Received data", "bytes", len(buf), "data", string(buf))
	kind, contents, err := c.decodeDocument(c.url.Location(), bytes.NewReader(buf))
	if err != nil {
		return fmt.Errorf("decode document: %w", err)
	}
	c.logger.Debug("Document decoded", "url", c.url, "kind", kind)
	c.kind = kind
	c.objectsTree = contents
	return nil
}

func (c *Document) readFile(u *jsonpointer.JSONPointer, locator documentLocator) ([]byte, error) {
	t := time.Now()
	defer func() {
		c.logger.Debug("File locator finished", "url", u, "duration", time.Since(t))
	}()

	data, err := locator.Locate(u)
	if err != nil {
		return nil, fmt.Errorf("locate document url %q: %w", u, err)
	}
	defer data.Close()

	return io.ReadAll(data)
}

// Compile runs the document compilation.
func (c *Document) Compile(ctx *compile.Context) error {
	ctx = ctx.WithResultsStore(c)
	c.logger.Debug("Compiling a document", "url", c.url, "kind", c.kind)
	c.logger.Trace("Compiling the root component", "url", c.url)
	if err := c.objectsTree.Compile(ctx); err != nil {
		return fmt.Errorf("root asyncapi component: %w", err)
	}
	c.logger.Trace("Compiling the nested components", "url", c.url)
	return WalkAndCompile(ctx, reflect.ValueOf(c.objectsTree))
}

func (c *Document) ExternalURLs() []*jsonpointer.JSONPointer {
	return c.externalRefs
}

func (c *Document) Stats() string {
	return fmt.Sprintf(
		"%s(%s): %d artifacts; %d external refs; %d promises; %d list promises",
		c.url, c.kind, len(c.Artifacts()), len(c.ExternalURLs()), len(c.Promises()), len(c.ListPromises()),
	)
}

func (c *Document) decodeDocument(filePath string, data io.ReadSeeker) (DocumentKind, compiledObject, error) {
	switch path.Ext(filePath) {
	case ".yaml", ".yml":
		c.logger.Debug("Found YAML file", "path", filePath)
		kind, contents, err := guessDocumentKind(yaml.NewDecoder(data))
		if err != nil {
			return "", nil, fmt.Errorf("guess document kind: %w", err)
		}
		if _, err = data.Seek(0, io.SeekStart); err != nil {
			return "", nil, fmt.Errorf("file seek: %w", err)
		}
		err = yaml.NewDecoder(data).Decode(contents)
		return kind, contents, err
	case ".json":
		c.logger.Debug("Found JSON file", "path", filePath)
		kind, contents, err := guessDocumentKind(json.NewDecoder(data))
		if err != nil {
			return "", nil, fmt.Errorf("guess document kind: %w", err)
		}
		if _, err = data.Seek(0, io.SeekStart); err != nil {
			return "", nil, fmt.Errorf("file seek: %w", err)
		}
		err = json.NewDecoder(data).Decode(contents)
		return kind, contents, err
	}

	return "", nil, fmt.Errorf("unknown kind with file extension: %s", filePath)
}
