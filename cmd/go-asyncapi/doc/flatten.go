package doc

import (
	"bytes"
	"fmt"
	"os"
	"slices"
	"time"

	common2 "github.com/bdragon300/go-asyncapi/cmd/go-asyncapi/common"
	"github.com/bdragon300/go-asyncapi/internal/compiler"
	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/samber/lo"
)

type FlattenCmd struct {
	Document     string `arg:"positional,required" help:"AsyncAPI document file or URL" placeholder:"FILE"`
	Output       string `arg:"--output,-o" help:"File where to write the flattened document. By default, the original document is modified in-place" placeholder:"FILE"`
	ExternalRefs bool   `arg:"--external-refs" help:"Consider the referenced objects located in external documents"`
	RemoteRefs   bool   `arg:"--remote-refs" help:"Consider the referenced objects located in documents addressed by URLs"`

	Indent int `arg:"--indent" help:"Output document indentation width" placeholder:"SPACES"`

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
		return fmt.Errorf("%w: cannot rewrite a remote document in-place, use the --output option to specify the output file", common2.ErrInvalidCLIArgument)
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
	documents := map[string]*common2.DocumentTree{inputContents.AbsOriginDocumentPath().Location(): inputContents}
	if _, err = flattenNode(inputContents.RawNode, documents, locator, nil, cmdConfig); err != nil {
		return fmt.Errorf("flatten document: %w", err)
	}

	logger.Info("Writing flattened document", "file", outputPath)
	buf := bytes.NewBuffer(nil)
	ef, _, format := compiler.GuessDocumentFormat(outputPath, cmdConfig.Doc.Flatten.Indent)
	if format == "" {
		return fmt.Errorf("cannot determine the output document format: %q", outputPath)
	}
	if err = ef(buf).Encode(inputContents); err != nil {
		return fmt.Errorf("marshal document: %w", err)
	}
	if err = os.WriteFile(outputPath, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write output file %q: %w", outputPath, err)
	}

	return nil
}

func flattenNode(
	node *types.RawNode,
	documents map[string]*common2.DocumentTree,
	locator common2.DocumentLocator,
	visited []*types.RawNode,
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
			unresolvable := lo.SomeBy(asyncapiUnresolvableRefPaths(), func(p []string) bool {
				return len(p) == len(n.Path()) && lo.EveryBy(lo.Range(len(p)), func(i int) bool { return p[i] == "" || p[i] == n.Path()[i] })
			})
			if unresolvable {
				logger.Debug("Skipping resolving a $ref due to AsyncAPI v3 specification rules", "path", n)
				continue
			}

			v = append(v, n)
			ref, err := parseRefRawNode(n)
			if err != nil {
				logger.Warn("Failed to parse $ref, skipping", "path", n, "ref", ref, "error", err.Error())
			}
			targetNode, err := resolveRefNode(ref, n, documents, locator, cmdConfig.Doc.Flatten.ExternalRefs, cmdConfig.Doc.Flatten.RemoteRefs)
			if err != nil {
				logger.Warn(fmt.Sprintf("%s, skipping", err.Error()), "path", n)
				continue // Can't resolve the $ref, skip it
			}
			if slices.Contains(v, targetNode) {
				logger.Warn("Detected a $ref cycle, leaving the $ref unresolved", "path", n, "ref", ref)
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
		clog, err := flattenNode(n, documents, locator, v, cmdConfig)
		if err != nil {
			return nil, err
		}
		changeLog = append(changeLog, clog...)
	}

	return changeLog, nil
}
