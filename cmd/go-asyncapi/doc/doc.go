package doc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"unicode"

	"github.com/bdragon300/go-asyncapi/cmd/go-asyncapi/common"
	"github.com/bdragon300/go-asyncapi/internal/compiler"
	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/gobwas/glob"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

type anyEncoder interface {
	Encode(v any) error
}

type Cmd struct {
	Cp          *CpCmd          `arg:"subcommand:cp" help:"Copy nodes between AsyncAPI documents."`
	Mv          *MvCmd          `arg:"subcommand:mv" help:"Move nodes between AsyncAPI documents."`
	Validate    *ValidateCmd    `arg:"subcommand:validate" help:"Validate AsyncAPI documents against the AsyncAPI JSON Schema."`
	Flatten     *FlattenCmd     `arg:"subcommand:flatten" help:"Flatten an AsyncAPI document by inlining all $refs with the nodes they point to."`
	GenExamples *GenExamplesCmd `arg:"subcommand:gen-examples" help:"Generate examples for messages, message traits and schemas in an AsyncAPI document."`
	Inspect     *InspectCmd     `arg:"subcommand:inspect" help:"Inspect the AsyncAPI entities"`
	Tree        *TreeCmd        `arg:"subcommand:tree" help:"Show the documents tree"`
}

type changeLogEntry struct {
	From, To *jsonpointer.JSONPointer
	Move     bool
}

func CliDoc(cmd *Cmd, globalConfig common2.ToolConfig) error {
	logger := log.GetLogger("")
	cmdConfig, err := cliConfig(globalConfig, cmd)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	logger.TraceYAML("Merged config", cmdConfig)

	switch {
	case cmd.Cp != nil:
		return cliCp(cmd.Cp, cmdConfig)
	case cmd.Mv != nil:
		return cliMv(cmd.Mv, cmdConfig)
	case cmd.Validate != nil:
		return cliValidate(cmd.Validate, cmdConfig)
	case cmd.Flatten != nil:
		return cliFlatten(cmd.Flatten, cmdConfig)
	case cmd.GenExamples != nil:
		return cliGenExamples(cmd.GenExamples, cmdConfig)
	case cmd.Inspect != nil:
		return cliInspect(cmd.Inspect, cmdConfig)
	case cmd.Tree != nil:
		return cliTree(cmd.Tree, cmdConfig)
	}
	return fmt.Errorf("%w: unknown doc subcommand", common2.ErrInvalidCLIArgument)
}

func cliConfig(globalConfig common2.ToolConfig, cmd *Cmd) (common2.ToolConfig, error) {
	res := globalConfig

	cmdValidate := lo.FromPtr(cmd.Validate)
	cmdFlatten := lo.FromPtr(cmd.Flatten)
	cmdCp := lo.FromPtr(cmd.Cp)
	cmdMv := lo.FromPtr(cmd.Mv)
	cmdGenExamples := lo.FromPtr(cmd.GenExamples)
	cmdInspect := lo.FromPtr(cmd.Inspect)
	cmdTree := lo.FromPtr(cmd.Tree)

	res.Doc.Validate.Schema = common2.Coalesce(cmdValidate.Schema, globalConfig.Doc.Validate.Schema)
	res.Doc.Flatten.OutputFile = common2.Coalesce(cmdFlatten.Output, globalConfig.Doc.Flatten.OutputFile)
	res.Doc.Flatten.ExternalRefs = common2.Coalesce(cmdFlatten.ExternalRefs, globalConfig.Doc.Flatten.ExternalRefs)
	res.Doc.Flatten.RemoteRefs = common2.Coalesce(cmdFlatten.RemoteRefs, globalConfig.Doc.Flatten.RemoteRefs)
	res.Doc.Flatten.Indent = common2.Coalesce(cmdFlatten.Indent, globalConfig.Doc.Flatten.Indent)
	res.Doc.Flatten.Format = common2.Coalesce(cmdFlatten.Format, globalConfig.Doc.Flatten.Format)
	if cmd.Flatten != nil {
		res.Locator.AllowRemoteReferences = common2.Coalesce(cmdFlatten.RemoteRefs, res.Doc.Flatten.RemoteRefs)
		res.Locator.Command = common2.Coalesce(cmdFlatten.LocatorCommand, globalConfig.Locator.Command)
		res.Locator.Timeout = common2.Coalesce(cmdFlatten.LocatorTimeout, globalConfig.Locator.Timeout)
		res.Locator.RootDirectory = common2.Coalesce(cmdFlatten.LocatorRootDir, globalConfig.Locator.RootDirectory)
	}
	res.Doc.Cp.AutoCreate = common2.Coalesce(cmdCp.AutoCreate, globalConfig.Doc.Cp.AutoCreate)
	res.Doc.Cp.ShallowRefs = common2.Coalesce(cmdCp.ShallowRefs, globalConfig.Doc.Cp.ShallowRefs)
	res.Doc.Cp.FollowRefs = common2.Coalesce(cmdCp.FollowRefs, globalConfig.Doc.Cp.FollowRefs)
	res.Doc.Cp.Headless = common2.Coalesce(cmdCp.Headless, globalConfig.Doc.Cp.Headless)
	res.Doc.Cp.Force = common2.Coalesce(cmdCp.Force, globalConfig.Doc.Cp.Force)
	res.Doc.Cp.Interactive = common2.Coalesce(cmdCp.Interactive, globalConfig.Doc.Cp.Interactive)
	res.Doc.Cp.DisableRewriting = common2.Coalesce(cmdCp.DisableRewriting, globalConfig.Doc.Cp.DisableRewriting)
	res.Doc.Cp.Indent = common2.Coalesce(cmdCp.Indent, globalConfig.Doc.Cp.Indent)
	res.Doc.Cp.Format = common2.Coalesce(cmdCp.Format, globalConfig.Doc.Cp.Format)
	res.Doc.Mv.AutoCreate = common2.Coalesce(cmdMv.AutoCreate, globalConfig.Doc.Mv.AutoCreate)
	res.Doc.Mv.Link = common2.Coalesce(cmdMv.Link, globalConfig.Doc.Mv.Link)
	res.Doc.Mv.ShallowRefs = common2.Coalesce(cmdMv.ShallowRefs, globalConfig.Doc.Mv.ShallowRefs)
	res.Doc.Mv.FollowRefs = common2.Coalesce(cmdMv.FollowRefs, globalConfig.Doc.Mv.FollowRefs)
	res.Doc.Mv.Headless = common2.Coalesce(cmdMv.Headless, globalConfig.Doc.Mv.Headless)
	res.Doc.Mv.Force = common2.Coalesce(cmdMv.Force, globalConfig.Doc.Mv.Force)
	res.Doc.Mv.Interactive = common2.Coalesce(cmdMv.Interactive, globalConfig.Doc.Mv.Interactive)
	res.Doc.Mv.DisableRewriting = common2.Coalesce(cmdMv.DisableRewriting, globalConfig.Doc.Mv.DisableRewriting)
	res.Doc.Mv.Indent = common2.Coalesce(cmdMv.Indent, globalConfig.Doc.Mv.Indent)
	res.Doc.Mv.Format = common2.Coalesce(cmdMv.Format, globalConfig.Doc.Mv.Format)
	res.Doc.GenExamples.OutputFile = common2.Coalesce(cmdGenExamples.Output, globalConfig.Doc.GenExamples.OutputFile)
	res.Doc.GenExamples.OnlyMessages = common2.Coalesce(cmdGenExamples.OnlyMessages, globalConfig.Doc.GenExamples.OnlyMessages)
	res.Doc.GenExamples.OnlySchemas = common2.Coalesce(cmdGenExamples.OnlySchemas, globalConfig.Doc.GenExamples.OnlySchemas)
	res.Doc.GenExamples.Append = common2.Coalesce(cmdGenExamples.Append, globalConfig.Doc.GenExamples.Append)
	res.Doc.GenExamples.Count = common2.Coalesce(cmdGenExamples.Count, globalConfig.Doc.GenExamples.Count)
	res.Doc.GenExamples.DateFormat = common2.Coalesce(cmdGenExamples.DateFormat, globalConfig.Doc.GenExamples.DateFormat)
	res.Doc.GenExamples.TimeFormat = common2.Coalesce(cmdGenExamples.TimeFormat, globalConfig.Doc.GenExamples.TimeFormat)
	res.Doc.GenExamples.DateTimeFormat = common2.Coalesce(cmdGenExamples.DateTimeFormat, globalConfig.Doc.GenExamples.DateTimeFormat)
	res.Doc.GenExamples.Indent = common2.Coalesce(cmdGenExamples.Indent, globalConfig.Doc.GenExamples.Indent)
	res.Doc.GenExamples.Format = common2.Coalesce(cmdGenExamples.Format, globalConfig.Doc.GenExamples.Format)
	if cmd.GenExamples != nil {
		res.Locator.AllowRemoteReferences = common2.Coalesce(cmdGenExamples.AllowRemoteRefs, res.Doc.GenExamples.AllowRemoteReferences)
		res.Locator.Command = common2.Coalesce(cmdGenExamples.LocatorCommand, globalConfig.Locator.Command)
		res.Locator.Timeout = common2.Coalesce(cmdGenExamples.LocatorTimeout, globalConfig.Locator.Timeout)
		res.Locator.RootDirectory = common2.Coalesce(cmdGenExamples.LocatorRootDir, globalConfig.Locator.RootDirectory)
	}
	res.Doc.Inspect.Entities = common2.Coalesce(cmdInspect.Entities, globalConfig.Doc.Inspect.Entities)
	res.Doc.Inspect.Recursive = common2.Coalesce(cmdInspect.Recursive, globalConfig.Doc.Inspect.Recursive)
	res.Doc.Inspect.RecursiveExpand = common2.Coalesce(cmdInspect.RecursiveExpand, globalConfig.Doc.Inspect.RecursiveExpand)
	res.Doc.Inspect.Components = common2.Coalesce(cmdInspect.Components, globalConfig.Doc.Inspect.Components)
	res.Doc.Inspect.Main = common2.Coalesce(cmdInspect.Main, globalConfig.Doc.Inspect.Main)
	res.Doc.Inspect.FollowExternalRefs = common2.Coalesce(cmdInspect.FollowExternalRefs, globalConfig.Doc.Inspect.FollowExternalRefs)
	res.Doc.Inspect.AllowRemoteReferences = common2.Coalesce(cmdInspect.AllowRemoteRefs, globalConfig.Doc.Inspect.AllowRemoteReferences)
	res.Doc.Inspect.List = common2.Coalesce(cmdInspect.List, globalConfig.Doc.Inspect.List)
	res.Doc.Inspect.EntryStyle = common2.Coalesce(cmdInspect.EntryStyle, globalConfig.Doc.Inspect.EntryStyle)
	if cmd.Inspect != nil {
		res.Locator.AllowRemoteReferences = common2.Coalesce(cmdInspect.AllowRemoteRefs, res.Doc.Inspect.AllowRemoteReferences)
		res.Locator.Command = common2.Coalesce(cmdInspect.LocatorCommand, globalConfig.Locator.Command)
		res.Locator.Timeout = common2.Coalesce(cmdInspect.LocatorTimeout, globalConfig.Locator.Timeout)
		res.Locator.RootDirectory = common2.Coalesce(cmdInspect.LocatorRootDir, globalConfig.Locator.RootDirectory)
	}
	res.Doc.Tree.List = common2.Coalesce(cmdTree.List, globalConfig.Doc.Tree.List)
	if cmd.Tree != nil {
		res.Locator.AllowRemoteReferences = common2.Coalesce(cmdTree.AllowRemoteRefs, res.Doc.Tree.AllowRemoteReferences)
		res.Locator.Command = common2.Coalesce(cmdTree.LocatorCommand, globalConfig.Locator.Command)
		res.Locator.Timeout = common2.Coalesce(cmdTree.LocatorTimeout, globalConfig.Locator.Timeout)
		res.Locator.RootDirectory = common2.Coalesce(cmdTree.LocatorRootDir, globalConfig.Locator.RootDirectory)
	}

	return res, nil
}

// collectRefs recursively collects all n's children RawNodes (including n itself) that have a "$ref" key.
func collectRefs(n *types.RawNode) []*types.RawNode {
	if n == nil {
		return nil
	}
	if n.Kind() == types.RawNodeKindScalar {
		return nil
	}

	var res []*types.RawNode
	for _, v := range n.Entries() {
		res = append(res, collectRefs(v)...)
	}
	if n.Kind() == types.RawNodeKindObject && n.Has("$ref") {
		res = append(res, n)
	}

	return res
}

func parseRefRawNode(n *types.RawNode) (*jsonpointer.JSONPointer, error) {
	variant, _ := n.Get("$ref")
	if variant.Kind() != types.RawNodeKindScalar {
		return nil, fmt.Errorf("expected scalar node for $ref value, got %s", variant.Kind())
	}
	s, ok := variant.AsScalar().(string)
	if !ok {
		return nil, fmt.Errorf("expected string value for $ref, got %T", variant.AsScalar())
	}

	ref, err := jsonpointer.Parse(s)
	if err != nil {
		return nil, fmt.Errorf("parse $ref value: %w", err)
	}
	return ref, nil
}

func getDocumentEncoder(w io.Writer, format string, indent int) (anyEncoder, error) {
	// TODO: auto guess output format based on input files extensions
	logger := log.GetLogger("")

	var enc anyEncoder
	switch format {
	case "json":
		logger.Debug("Using JSON output format", "indent", indent)
		e := json.NewEncoder(w)
		e.SetIndent("", strings.Repeat(" ", indent))
		enc = e
	case "yaml":
		logger.Debug("Using YAML output format", "indent", indent)
		e := yaml.NewEncoder(w)
		e.SetIndent(indent)
		enc = e
	default:
		return nil, fmt.Errorf("%w: unknown output format %q", common2.ErrInvalidCLIArgument, format)
	}
	return enc, nil
}

func rewriteRefs(doc *common2.DocumentTree, locator common2.DocumentLocator, changeLog []changeLogEntry) {
	logger := log.GetLogger("")

	for _, r := range collectRefs(doc.RawNode) {
		logger.Debug("Processing $ref", "path", r.Path())
		if r.AbsOriginDocumentPath() == nil {
			logger.Warn("Found a $ref with empty metadata, this is a bug, skipping", "path", r.Path())
			continue
		}
		originPath := r.AbsOriginDocumentPath() // Document path where this $ref was imported from
		// Explicit document path where this $ref pointed to before been imported. For internal $ref, it's the file itself
		targetPath := originPath

		ref, err := parseRefRawNode(r)
		if err != nil {
			logger.Error("Failed to parse $ref, skipping", "path", r.Path(), "error", err)
			continue
		}
		logger.Trace("Parsed $ref", "path", r.Path(), "value", ref, "originDocument", originPath)
		if ref.Location() != "" {
			// Resolve location in external $ref relative to it's origin document location
			if targetPath, err = locator.ResolveURL(originPath, ref); err != nil {
				logger.Error("Failed to resolve $ref, skipping", "path", r.Path(), "value", ref, "error", err)
				continue
			}
		}

		logger.Trace("Rewriting $ref", "path", r.Path(), "value", ref, "originDocument", originPath, "referredDocument", targetPath)
		newRef := rewriteRef(ref, targetPath, originPath, doc, changeLog)
		if newRef == nil {
			continue
		}

		logger.Debug("Updating rewritten $ref", "path", r.Path(), "old", ref, "new", newRef)
		r.Set("$ref", types.NewScalarRawNode(r.Path(), newRef.String(), r.AbsOriginDocumentPath()))
	}
}

func rewriteRef(ref, targetPath, originPath *jsonpointer.JSONPointer, doc *common2.DocumentTree, changeLog []changeLogEntry) *jsonpointer.JSONPointer {
	logger := log.GetLogger("")

	newRef := *ref
	if len(ref.Pointer) == 0 {
		logger.Trace("$ref without pointer, skipping")
		return ref
	}
	resolveRelPath := func(dt *common2.DocumentTree, p *jsonpointer.JSONPointer) string {
		rel, err := filepath.Rel(path.Dir(dt.AbsOriginDocumentPath().Location()), absLocation(p))
		if err != nil {
			logger.Error("Failed to rewrite external location in $ref, leaving it as-is", "value", ref, "error", err)
			return ""
		}
		logger.Trace("Fixing $ref location to external document", "location", rel)
		return rel
	}

	change, inChangeLog := lo.Find(changeLog, func(c changeLogEntry) bool {
		logger.Trace("Checking changelog entry", "entry", c)
		if c.From == nil || len(c.From.Pointer) == 0 || c.To == nil || len(c.To.Pointer) == 0 {
			return false
		}

		pointerOk := lo.HasPrefix(ref.Pointer, c.From.Pointer)
		// Ref points to the old or new location of node (or its child)
		return (absLocation(c.From) == absLocation(targetPath) || absLocation(c.To) == absLocation(targetPath)) && pointerOk
	})

	// $ref was moved or copied to doc from another document
	immigratedRef := absLocation(originPath) != doc.AbsOriginDocumentPath().Location()

	if inChangeLog {
		logger.Trace("Found $ref in changelog", "entry", change, "ref", ref)
		// Keep the $ref's pointer if it's original for doc and points to a node that is still there (i.e. was copied somewhere).
		keepPointer := !change.Move && !immigratedRef && absLocation(targetPath) == absLocation(change.From)
		if !keepPointer {
			// $ref was not relocated, but the target node was relocated away from doc or another external document
			pathSuffix, _ := lo.CutPrefix(ref.Pointer, change.From.Pointer)
			newRef.Pointer = append(change.To.Pointer, pathSuffix...) // nolint:gocritic
			logger.Trace("Updating $ref pointer")
		}

		// Target node has been relocated, regardless of whether $ref was relocated or not
		switch {
		case absLocation(change.To) == doc.AbsOriginDocumentPath().Location():
			// Node has been relocated into doc
			logger.Trace("Node has been relocated into doc")
			newRef.FSPath = ""
			newRef.URI = nil
		case absLocation(change.From) == absLocation(change.To):
			// Node has only been renamed without relocation
			logger.Trace("Node has been renamed without relocation")
		case change.To.FSPath != "" && change.Move:
			// Node has been relocated away from doc or relocated between two external documents or relocated from external URL to external file path
			logger.Trace("Node has been relocated away from doc or relocated between two external documents or relocated from external URL to external file path")
			newRef.FSPath = resolveRelPath(doc, change.To)
			newRef.URI = nil
		case change.To.URI != nil && change.Move:
			// Node has been relocated away from doc or relocated between two external documents or relocated from external file path to external URL
			logger.Trace("Node has been relocated away from doc or relocated between two external documents or relocated from external file path to external URL")
			newRef.FSPath = ""
			newRef.URI = change.To.URI
		}

		logger.Trace("Rewrite result for $ref", "old", ref, "new", newRef)
		return &newRef
	}

	switch {
	case !immigratedRef:
		// Neither $ref nor the node it pointed to was relocated, so we don't need to rewrite it
		logger.Trace("Neither $ref nor the node it pointed to was relocated, skipping")
	case absLocation(originPath) == absLocation(targetPath):
		// Local $ref was relocated into doc (because we see it), but the node it pointed to was not relocated
		logger.Trace("Local $ref was relocated into doc, but the node it pointed to was not relocated")
		newRef.FSPath = resolveRelPath(doc, originPath)
	case absLocation(targetPath) == doc.AbsOriginDocumentPath().Location():
		// $ref pointing to doc was relocated into doc (because we see it), but the node it pointed to was not relocated
		logger.Trace("$ref pointing to doc was relocated into doc, but the node it pointed to was not relocated")
		newRef.FSPath = ""
		newRef.URI = nil
	case ref.FSPath != "":
		// $ref pointing to external document file path was relocate into doc (because we see it), but the node it pointed to was not relocated
		logger.Trace("$ref pointing to external document file path was relocated into doc, but the node it pointed to was not relocated")
		newRef.FSPath = resolveRelPath(doc, targetPath)
	case ref.URI != nil:
		// $ref pointing to external document URL was relocated into doc (because we see it), but the node it pointed to was not relocated
		logger.Trace("$ref pointing to external document URL was relocated into doc, but the node it pointed to was not relocated")
	}

	logger.Trace("Rewrite result for $ref", "old", ref, "new", newRef)
	return &newRef
}

// absLocation returns the absolute path to document if p is a file path. Otherwise, if p is an URL, it returns the URL as is.
func absLocation(p *jsonpointer.JSONPointer) string {
	if p.FSPath != "" {
		return lo.Must(filepath.Abs(p.FSPath))
	}
	return p.Location()
}

func loadDocument(inputDoc *jsonpointer.JSONPointer, locator common2.DocumentLocator) (*common2.DocumentTree, error) {
	absInputPath := lo.Must(jsonpointer.Parse(absLocation(inputDoc)))
	inputContents := common2.NewDocumentTree(absInputPath)
	buf, newDecoder, err := compiler.ReadDocument(inputDoc, locator, log.GetLogger(""))
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	hasContents := bytes.ContainsFunc(buf, func(r rune) bool {
		return unicode.IsDigit(r) || unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsPunct(r)
	})
	if !hasContents {
		return inputContents, nil
	}

	if err = newDecoder(bytes.NewReader(buf)).Decode(&inputContents); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return inputContents, nil
}

func resolveRefNode(ref *jsonpointer.JSONPointer, refNode *types.RawNode, docs map[string]*common2.DocumentTree, locator common2.DocumentLocator, external, remote bool) (*types.RawNode, error) {
	logger := log.GetLogger("")
	var err error

	// Figure out which document the $ref points to. A local $ref (no location part) points to the document it was read
	// from. An external $ref is resolved relative to its origin document using the locator.
	targetDoc := refNode.AbsOriginDocumentPath()
	if ref.Location() != "" {
		if ref.FSPath != "" && !external {
			return nil, fmt.Errorf("objects from external documents are disabled, use the --with-external flag to enable")
		}
		// TODO: fix error messages snice this function is shared
		if ref.URI != nil && !remote {
			return nil, fmt.Errorf("objects from remote documents are disabled, use the --with-remote flag to enable")
		}
		if targetDoc, err = locator.ResolveURL(refNode.AbsOriginDocumentPath(), ref); err != nil {
			return nil, fmt.Errorf("resolve $ref location %q: %w", ref.Location(), err)
		}
	}

	targetAbsLoc := absLocation(targetDoc)
	document, ok := docs[targetAbsLoc]
	if !ok {
		logger.Debug("Loading the referenced document", "location", targetAbsLoc)
		if document, err = loadDocument(targetDoc, locator); err != nil {
			return nil, fmt.Errorf("load the referenced document: %w", err)
		}
		docs[targetAbsLoc] = document
	}

	n := document.GetByPath(ref.Pointer)
	if n == nil {
		return n, fmt.Errorf("pointer %q not found in document %q", ref.PointerString(), targetAbsLoc)
	}
	return n, nil
}

type cliPattern struct {
	*jsonpointer.JSONPointer
	pattern glob.Glob
}

func (c cliPattern) MatchPath(p []string) bool {
	if len(c.Pointer) == 0 || slices.Equal(p, c.Pointer) {
		return true
	}
	r := c.pattern.Match(strings.Join(p, "/"))
	return r
}

func parseCliPattern(arg string) (cliPattern, error) {
	p, err := jsonpointer.Parse(arg)
	if err != nil {
		return cliPattern{}, fmt.Errorf("parse %q: %w", arg, err)
	}
	gl, err := glob.Compile(strings.Join(p.Pointer, "/"), '/')
	if err != nil {
		return cliPattern{}, fmt.Errorf("compile glob pattern %q: %w", arg, err)
	}
	return cliPattern{JSONPointer: p, pattern: gl}, nil
}

func findNodesByPattern(node *types.RawNode, pattern cliPattern) []*types.RawNode {
	if node == nil {
		return nil
	}

	var res []*types.RawNode
	if len(node.Path()) > 0 && pattern.MatchPath(node.Path()) {
		res = append(res, node)
	}

	if node.Kind() == types.RawNodeKindObject || node.Kind() == types.RawNodeKindArray {
		for _, e := range node.Entries() {
			res = append(res, findNodesByPattern(e, pattern)...)
		}
	}
	return res
}
