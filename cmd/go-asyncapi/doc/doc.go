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

	"github.com/bdragon300/go-asyncapi/cmd/go-asyncapi/common"
	"github.com/bdragon300/go-asyncapi/internal/compiler"
	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

type anyEncoder interface {
	Encode(v any) error
}

type Cmd struct {
	Merge    *MergeCmd    `arg:"subcommand:merge" help:"Merge multiple AsyncAPI documents into one."`
	Unmerge  *UnmergeCmd  `arg:"subcommand:unmerge" help:"Unmerge an AsyncAPI document by relocating the objects to another document."`
	Validate *ValidateCmd `arg:"subcommand:validate" help:"Validate AsyncAPI documents against the AsyncAPI JSON Schema."`
	Indent   int          `arg:"--indent" help:"Output document indentation width" placeholder:"SPACES"`
	Format   string       `arg:"--format,-f" help:"Output format. Possible values: yaml, json" placeholder:"FORMAT"`
}

func newDocumentTree(originDocument *jsonpointer.JSONPointer) *documentTree {
	return &documentTree{RawNode: types.NewEmptyRawNode(types.RawNodeKindObject, nil, originDocument)}
}

type documentTree struct {
	*types.RawNode
}

// CollectRefs collects pointers to all object RawNodes in the document that have a "$ref" key.
func (d documentTree) CollectRefs() []*types.RawNode {
	rootKeys, componentsKeys := d.MergeableMapPaths()
	if d.IsZero() {
		return nil
	}
	res := lo.FlatMap(rootKeys, func(key string, _ int) []*types.RawNode {
		node, _ := d.Get(key)
		return collectRefs(node)
	})

	if comp, ok := d.Get("components"); ok {
		res = append(res, lo.FlatMap(componentsKeys, func(key string, _ int) []*types.RawNode {
			if comp.IsZero() {
				return nil
			}
			node, _ := comp.Get(key)
			return collectRefs(node)
		})...)
	}

	return res
}

func (d documentTree) MergeableMaps() []*types.RawNode {
	rootKeys, componentsKeys := d.MergeableMapPaths()
	if d.IsZero() {
		return make([]*types.RawNode, len(rootKeys)+len(componentsKeys))
	}
	res := lo.Map(rootKeys, func(key string, _ int) *types.RawNode {
		node, _ := d.Get(key)
		return node
	})
	if comp, ok := d.Get("components"); ok {
		res = append(res, lo.Map(componentsKeys, func(key string, _ int) *types.RawNode {
			if comp.IsZero() {
				return nil
			}
			node, _ := comp.Get(key)
			return node
		})...)
	}

	return res
}

func (d documentTree) MergeableMapPaths() ([]string, []string) {
	rootKeys := []string{"servers", "channels", "operations"}
	componentsKeys := []string{
		"schemas",
		"servers", "channels", "operations", "messages",
		"securitySchemes", "serverVariables", "parameters", "correlationIds", "replies", "replyAddresses", "externalDocs", "tags",
		"operationTraits", "messageTraits",
		"serverBindings", "channelBindings", "operationBindings", "messageBindings",
	}
	return rootKeys, componentsKeys
}

func (d documentTree) EssentialNodesPaths() []string {
	return []string{"asyncapi", "info", "defaultContentType"}
}

type changeLogEntry struct {
	From, To *jsonpointer.JSONPointer
	Move     bool
}

func CliDoc(cmd *Cmd, globalConfig common2.ToolConfig) error {
	cmdConfig, err := cliConfig(globalConfig, cmd)
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}

	switch {
	case cmd.Merge != nil:
		return cliMerge(cmd.Merge, cmdConfig)
	case cmd.Unmerge != nil:
		return cliUnmerge(cmd.Unmerge, cmdConfig)
	case cmd.Validate != nil:
		return cliValidate(cmd.Validate, cmdConfig)
	}
	return fmt.Errorf("%w: unknown doc subcommand", common2.ErrWrongCliArgs)
}

func cliConfig(globalConfig common2.ToolConfig, cmd *Cmd) (common2.ToolConfig, error) {
	res := globalConfig

	cmdMerge := lo.FromPtr(cmd.Merge)
	cmdUnmerge := lo.FromPtr(cmd.Unmerge)
	cmdValidate := lo.FromPtr(cmd.Validate)

	res.Doc.Indent = common2.Coalesce(cmd.Indent, globalConfig.Doc.Indent)
	res.Doc.Format = common2.Coalesce(cmd.Format, globalConfig.Doc.Format)
	res.Doc.Merge.OutputFile = common2.Coalesce(cmdMerge.Output, globalConfig.Doc.Merge.OutputFile)
	res.Doc.Merge.Strategy = common2.Coalesce(cmdMerge.Strategy, globalConfig.Doc.Merge.Strategy)
	res.Doc.Merge.DisableRewriting = common2.Coalesce(cmdMerge.DisableRewriting, globalConfig.Doc.Merge.DisableRewriting)
	res.Doc.Unmerge.UnmergeTo = common2.Coalesce(cmdUnmerge.UnmergeTo, globalConfig.Doc.Unmerge.UnmergeTo)
	res.Doc.Unmerge.Output = common2.Coalesce(cmdUnmerge.Output, globalConfig.Doc.Unmerge.Output)
	res.Doc.Unmerge.IncludeDeps = common2.Coalesce(cmdUnmerge.IncludeDeps, globalConfig.Doc.Unmerge.IncludeDeps)
	res.Doc.Unmerge.DuplicateObjects = common2.Coalesce(cmdUnmerge.DuplicateObjects, globalConfig.Doc.Unmerge.DuplicateObjects)
	res.Doc.Unmerge.DisableRewriting = common2.Coalesce(cmdUnmerge.DisableRewriting, globalConfig.Doc.Unmerge.DisableRewriting)
	res.Doc.Unmerge.NoInteractive = common2.Coalesce(cmdUnmerge.NoInteractive, globalConfig.Doc.Unmerge.NoInteractive)
	res.Doc.Validate.Schema = common2.Coalesce(cmdValidate.Schema, globalConfig.Doc.Validate.Schema)

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

func getDocumentEncoder(w io.Writer, cmdConfig common2.ToolConfig) (anyEncoder, error) {
	logger := log.GetLogger("")

	var enc anyEncoder
	switch cmdConfig.Doc.Format {
	case "json":
		logger.Debug("Using JSON output format", "indent", cmdConfig.Doc.Indent)
		e := json.NewEncoder(w)
		e.SetIndent("", strings.Repeat(" ", cmdConfig.Doc.Indent))
		enc = e
	case "yaml":
		logger.Debug("Using YAML output format", "indent", cmdConfig.Doc.Indent)
		e := yaml.NewEncoder(w)
		e.SetIndent(cmdConfig.Doc.Indent)
		enc = e
	default:
		return nil, fmt.Errorf("%w: unknown output format %q", common2.ErrWrongCliArgs, cmdConfig.Doc.Format)
	}
	return enc, nil
}

func rewriteRefs(doc *documentTree, locator common2.DocumentLocator, changeLog []changeLogEntry) {
	logger := log.GetLogger("")

	for _, r := range doc.CollectRefs() {
		logger.Debug("Processing $ref object", "path", r.Path())
		if r.OriginDocument() == nil {
			logger.Warn("Found a $ref with empty metadata, this is a bug, skipping", "path", r.Path())
			continue
		}
		originPath := r.OriginDocument() // Document path where this $ref was imported from
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

		logger.Debug("Rewriting $ref", "path", r.Path(), "value", ref, "originDocument", originPath, "referredDocument", targetPath)
		newRef := rewriteRef(ref, targetPath, originPath, doc, changeLog)
		if newRef == nil {
			continue
		}

		logger.Debug("Updating $ref", "path", r.Path(), "old", ref, "new", newRef)
		r.Set("$ref", types.NewScalarRawNode(r.Path(), newRef.String(), r.OriginDocument()))
	}
}

func rewriteRef(ref, targetPath, originPath *jsonpointer.JSONPointer, doc *documentTree, changeLog []changeLogEntry) *jsonpointer.JSONPointer {
	logger := log.GetLogger("")
	newRef := *ref
	// TODO: logging messages
	if len(ref.Pointer) == 0 {
		logger.Debug("-> $ref without pointer, skipping")
		return ref
	}
	resolveRelPath := func(dt *documentTree, p *jsonpointer.JSONPointer) string {
		rel, err := filepath.Rel(path.Dir(absLocation(dt.OriginDocument())), absLocation(p))
		if err != nil {
			logger.Error("Failed to rewrite external location in $ref, leaving it as-is", "value", ref, "error", err)
			return ""
		}
		logger.Debug("-> Fixing $ref location to external document", "location", rel)
		return rel
	}

	change, inChangeLog := lo.Find(changeLog, func(c changeLogEntry) bool {
		logger.Trace("-> Checking changelog entry", "entry", c)
		if c.From == nil || len(c.From.Pointer) == 0 || c.To == nil || len(c.To.Pointer) == 0 {
			return false
		}

		pointerOk := lo.HasPrefix(ref.Pointer, c.From.Pointer)
		// Ref points to the old or new location of object (or its child)
		ok := (absLocation(c.From) == absLocation(targetPath) || absLocation(c.To) == absLocation(targetPath)) && pointerOk
		if !ok {
			logger.Trace("--> No match")
		}
		return ok
	})

	if inChangeLog && !slices.Equal(change.From.Pointer, change.To.Pointer) {
		// Replace the pointer prefix if it has been changed in changelog (object rename)
		pathSuffix, _ := lo.CutPrefix(ref.Pointer, change.From.Pointer)
		newRef.Pointer = append(change.To.Pointer, pathSuffix...) // nolint:gocritic
	}

	if inChangeLog {
		// Target object has been moved or copied, regardless of whether $ref was moved or not
		switch {
		case absLocation(change.To) == absLocation(doc.OriginDocument()):
			// Object has been moved or copied to doc
			newRef.FSPath = ""
			newRef.URI = nil
		case absLocation(change.From) == absLocation(change.To):
			// Object has only been renamed in the same document
		case change.To.FSPath != "" && change.Move:
			// Object has been moved away from doc or moved between two external documents or moved from external URL to external file path
			newRef.FSPath = resolveRelPath(doc, change.To)
			newRef.URI = nil
		case change.To.URI != nil && change.Move:
			// Object has been moved away from doc or moved between two external documents or moved from external file path to external URL
			newRef.FSPath = ""
			newRef.URI = change.To.URI
		}

		return &newRef
	}

	// Target object has not been moved from its origin
	refWasMoved := absLocation(originPath) != absLocation(doc.OriginDocument())
	switch {
	case !refWasMoved:
		// Nor $ref nor the object it pointed to was moved, so we don't need to rewrite it
		logger.Debug("-> $ref was not moved, skipping")
	case absLocation(originPath) == absLocation(targetPath):
		// Local $ref was moved to doc (because we see it), but the object it pointed to was not moved
		newRef.FSPath = resolveRelPath(doc, originPath)
	case absLocation(targetPath) == absLocation(doc.OriginDocument()):
		// $ref pointing to doc was moved to doc (because we see it), but the object it pointed to was not moved
		newRef.FSPath = ""
		newRef.URI = nil
	case ref.FSPath != "":
		// $ref pointing to external document file path was moved to doc (because we see it), but the object it pointed to was not moved
		newRef.FSPath = resolveRelPath(doc, targetPath)
	case ref.URI != nil:
		// $ref pointing to external document URL was moved to doc (because we see it), but the object it pointed to was not moved
	}

	return &newRef
}

// absLocation returns the absolute path to document if p is a file path. Otherwise, if p is an URL, it returns the URL as is.
func absLocation(p *jsonpointer.JSONPointer) string {
	if p.FSPath != "" {
		return lo.Must(filepath.Abs(p.FSPath))
	}
	return p.Location()
}

func relocateEssentialNodes(sourceDoc, destDoc *documentTree) []changeLogEntry {
	logger := log.GetLogger("")

	// Relocate essential nodes if they are not present in destination document
	var changeLog []changeLogEntry
	for _, p := range sourceDoc.EssentialNodesPaths() {
		if destDoc.Has(p) {
			continue
		}
		if sNode, ok := sourceDoc.Get(p); ok {
			logger.Debug("Relocating object", "path", sNode.Path())
			dNode := sNode.DeepCopy()
			destDoc.Set(p, dNode)
			changeLog = append(changeLog, changeLogEntry{
				From: lo.ToPtr(sourceDoc.OriginDocument().Join(p)),
				To:   lo.ToPtr(destDoc.OriginDocument().Join(p)),
				Move: false,
			})
		}
	}
	return changeLog
}

func loadDocument(inputDoc *jsonpointer.JSONPointer, locator common2.DocumentLocator) (*documentTree, error) {
	logger := log.GetLogger("")

	absInputPath := lo.Must(jsonpointer.Parse(absLocation(inputDoc)))
	inputContents := newDocumentTree(absInputPath)
	buf, newDecoder, err := compiler.ReadDocument(inputDoc, locator, logger)
	if err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	if err = newDecoder(bytes.NewReader(buf)).Decode(&inputContents); err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return inputContents, nil
}
