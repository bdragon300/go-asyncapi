package doc

import (
	"bytes"
	"fmt"
	"os"
	"slices"
	"time"

	common2 "github.com/bdragon300/go-asyncapi/cmd/go-asyncapi/common"
	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/samber/lo"
)

type FlattenCmd struct {
	Document     string `arg:"positional,required" help:"AsyncAPI document file or URL" placeholder:"FILE"`
	Output       string `arg:"--output,-o" help:"File where to write the flattened document. By default, the original document is modified in-place" placeholder:"FILE"`
	WithExternal bool   `arg:"--with-external" help:"Consider the referenced objects located in external documents"`
	WithRemote   bool   `arg:"--with-remote" help:"Consider the referenced objects located in documents addressed by URLs"`

	LocatorRootDir string        `arg:"--locator-root-dir" help:"Root directory to search the documents" placeholder:"PATH"`
	LocatorTimeout time.Duration `arg:"--locator-timeout" help:"Timeout for locator to read a document. Format: 30s, 2m, etc." placeholder:"DURATION"`
	LocatorCommand string        `arg:"--locator-command" help:"Custom locator command to use instead of built-in locator" placeholder:"COMMAND"`
}

func cliFlatten(cmd *FlattenCmd, cmdConfig common2.ToolConfig) error {
	logger := log.GetLogger("")

	inputDoc, err := jsonpointer.Parse(cmd.Document)
	if err != nil {
		return fmt.Errorf("parse document url %q: %w", cmd.Document, err)
	}

	// A remote document cannot be rewritten in-place, so the user must provide an explicit output file for it.
	outputPath := cmdConfig.Doc.Flatten.OutputFile
	if inputDoc.URI != nil && outputPath == "" {
		return fmt.Errorf("%w: cannot rewrite a remote document in-place, use the --output option to specify the output file", common2.ErrWrongCliArgs)
	}
	if outputPath == "" {
		outputPath = cmd.Document // Rewrite the original document in-place
	}
	outputDoc, err := jsonpointer.Parse(outputPath)
	if err != nil {
		return fmt.Errorf("parse output path as url: %w", err)
	}
	if outputDoc.URI != nil {
		return fmt.Errorf("writing to documents by URL is not supported, please provide a file path")
	}

	logger.Debug("Loading document", "url", inputDoc)
	locator := common2.GetLocator(cmdConfig)
	inputContents, err := loadDocument(inputDoc, locator)
	if err != nil {
		return fmt.Errorf("load document: %w", err)
	}

	logger.Debug("Flattening the document", "url", inputDoc)
	documents := map[string]*documentTree{inputContents.AbsOriginDocumentPath().Location(): inputContents}
	if _, err = flattenNode(inputContents.RawNode, documents, locator, nil, inputContents.UnresolvableRefPaths(), cmdConfig); err != nil {
		return fmt.Errorf("flatten document: %w", err)
	}

	logger.Info("Writing flattened document", "file", outputPath)
	buf := bytes.NewBuffer(nil)
	enc, err := getDocumentEncoder(buf, cmdConfig)
	if err != nil {
		return fmt.Errorf("get encoder: %w", err)
	}
	if err = enc.Encode(inputContents); err != nil {
		return fmt.Errorf("marshal document: %w", err)
	}
	if err = os.WriteFile(outputPath, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write output file %q: %w", outputPath, err)
	}

	return nil
}

func flattenNode(
	node *types.RawNode,
	documents map[string]*documentTree,
	locator common2.DocumentLocator,
	visited []*types.RawNode,
	unresolvablePaths [][]string,
	cmdConfig common2.ToolConfig,
) ([]changeLogEntry, error) {
	logger := log.GetLogger("")

	if node == nil || node.Kind() == types.RawNodeKindScalar {
		return nil, nil
	}
	visited = append(visited, node)

	type entry struct {
		key  any
		node *types.RawNode
	}
	var snapshot []entry
	for k, v := range node.Entries() {
		snapshot = append(snapshot, entry{key: k, node: v})
	}

	var changeLog []changeLogEntry
	for _, e := range snapshot {
		n := e.node
		if n.Kind() == types.RawNodeKindScalar {
			continue
		}
		v := visited

		if n.Kind() == types.RawNodeKindObject && n.Has("$ref") {
			unresolvable := lo.SomeBy(unresolvablePaths, func(p []string) bool {
				return len(p) == len(n.Path()) && lo.EveryBy(lo.Range(len(p)), func(i int) bool { return p[i] == "" || p[i] == n.Path()[i] })
			})
			if unresolvable {
				logger.Debug("Skipping resolving a $ref due to AsyncAPI v3 specification rules", "path", n.AbsPointerString())
				continue
			}

			v = append(v, n)
			refPointer, targetNode, resolved := resolveRefNode(n, documents, locator, cmdConfig)
			if !resolved || targetNode == nil {
				continue // Can't resolve the $ref, skip it
			}
			if slices.Contains(v, targetNode) {
				logger.Warn("Detected a $ref cycle, leaving the $ref unresolved", "path", n.AbsPointerString(), "ref", refPointer)
				continue
			}
			node.Set(e.key, targetNode)
			changeLog = append(changeLog, changeLogEntry{
				From: lo.ToPtr(targetNode.AbsOriginDocumentPath().Join(targetNode.Path()...)),
				To:   lo.ToPtr(node.AbsOriginDocumentPath().Join(n.Path()...)),
				Move: false,
			})
			n = targetNode
		}

		// Recursively flatten the descendants, including the newly inlined ones, to resolve the nested $refs.
		clog, err := flattenNode(n, documents, locator, v, unresolvablePaths, cmdConfig)
		if err != nil {
			return nil, err
		}
		changeLog = append(changeLog, clog...)
	}

	return changeLog, nil
}

func resolveRefNode(node *types.RawNode, documents map[string]*documentTree, locator common2.DocumentLocator, cmdConfig common2.ToolConfig) (*jsonpointer.JSONPointer, *types.RawNode, bool) {
	logger := log.GetLogger("")

	if node.AbsOriginDocumentPath() == nil {
		logger.Warn("Found a $ref with empty metadata, this is a bug, skipping", "path", node.AbsPointerString())
		return nil, node, false
	}

	ref, err := parseRefRawNode(node)
	if err != nil {
		logger.Error("Failed to parse $ref, skipping", "path", node.AbsPointerString(), "error", err)
		return nil, node, false
	}
	logger.Debug("Resolving $ref", "path", node.AbsPointerString(), "ref", ref)

	// Figure out which document the $ref points to. A local $ref (no location part) points to the document it was read
	// from. An external $ref is resolved relative to its origin document using the locator.
	targetDoc := node.AbsOriginDocumentPath()
	if ref.Location() != "" {
		if ref.FSPath != "" && !cmdConfig.Doc.Flatten.WithExternal {
			logger.Warn("Inlining the objects from external documents are disabled, use the --with-external flag to enable", "path", node.AbsPointerString(), "ref", ref)
			return ref, node, false
		}
		if ref.URI != nil && !cmdConfig.Doc.Flatten.WithRemote {
			logger.Warn("Inlining the objects from remote documents are disabled, use the --with-remote flag to enable", "path", node.AbsPointerString(), "ref", ref)
			return ref, node, false
		}
		if targetDoc, err = locator.ResolveURL(node.AbsOriginDocumentPath(), ref); err != nil {
			logger.Error("Failed to resolve the $ref, skipping", "path", node.AbsPointerString(), "ref", ref, "error", err)
			return ref, node, false
		}
	}

	targetAbsLoc := absLocation(targetDoc)
	document, ok := documents[targetAbsLoc]
	if !ok {
		logger.Debug("Loading the referenced document", "location", targetAbsLoc)
		if document, err = loadDocument(targetDoc, locator); err != nil {
			logger.Error("Failed to load the document by $ref, skipping", "path", node.AbsPointerString(), "ref", ref, "error", err)
			return ref, node, false
		}
		documents[targetAbsLoc] = document
	}

	n := document.GetByPath(ref.Pointer)
	if n == nil {
		logger.Error("Failed to resolve $ref, skipping", "path", node.AbsPointerString())
	}
	return ref, n, n != nil
}
