package doc

import (
	"bytes"
	"fmt"
	"os"
	"slices"
	"strings"
	"time"

	common2 "github.com/bdragon300/go-asyncapi/cmd/go-asyncapi/common"
	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/bdragon300/go-asyncapi/internal/types"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/samber/lo"
)

type GenExamplesCmd struct {
	Location string `arg:"positional,required" help:"Document file or URL, optionally with a JSON pointer to a target subtree. Format: file.{yaml|yml|json}[#/path/to/node]" placeholder:"LOCATION"`
	Output   string `arg:"--output,-o" help:"File where to write the result. By default, the original document is modified in-place" placeholder:"FILE"`

	Append          bool `arg:"--append" help:"Append generated examples to entities that already have examples"`
	Count           int  `arg:"--count" help:"Number of examples to generate for each entity" placeholder:"COUNT"`
	OnlyMessages    bool `arg:"--only-messages" help:"Generate examples for messages only"`
	OnlySchemas     bool `arg:"--only-schemas" help:"Generate examples for schemas only (including nested ones in other entities)"`
	AllowRemoteRefs bool `arg:"--allow-remote-refs" help:"Allow to fetch the documents from remote hosts"`

	DateFormat     string `arg:"--date-format" help:"Go date format to use in 'date' fields. See: https://pkg.go.dev/time#pkg-constants" placeholder:"FORMAT_STRING"`
	TimeFormat     string `arg:"--time-format" help:"Go time format to use in 'time' fields. See: https://pkg.go.dev/time#pkg-constants" placeholder:"FORMAT_STRING"`
	DateTimeFormat string `arg:"--date-time-format" help:"Go date-time format to use in 'date-time' fields. See: https://pkg.go.dev/time#pkg-constants" placeholder:"FORMAT_STRING"`

	LocatorRootDir string        `arg:"--locator-root-dir" help:"Root directory to search the documents" placeholder:"PATH"`
	LocatorTimeout time.Duration `arg:"--locator-timeout" help:"Timeout for locator to read a document. Format: 30s, 2m, etc." placeholder:"DURATION"`
	LocatorCommand string        `arg:"--locator-command" help:"Custom locator command to use instead of built-in locator" placeholder:"COMMAND"`
}

func cliGenExamples(cmd *GenExamplesCmd, cmdConfig common2.ToolConfig) error {
	logger := log.GetLogger("")

	inputDoc, err := jsonpointer.Parse(cmd.Location)
	if err != nil {
		return fmt.Errorf("parse location %q: %w", cmd.Location, err)
	}

	// A remote document cannot be rewritten in-place, so the user must provide an explicit output file for it.
	outputPath := cmdConfig.Doc.GenExamples.OutputFile
	if outputPath == "" {
		if inputDoc.URI != nil {
			return fmt.Errorf("%w: cannot rewrite a remote document in-place, use the --output option to specify the output file", common2.ErrInvalidCLIArgument)
		}
		outputPath = inputDoc.Location() // Rewrite the original document in-place
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
	doc, err := loadDocument(inputDoc, locator)
	if err != nil {
		return fmt.Errorf("load document: %w", err)
	}

	// If a JSON pointer was given, only the addressed subtree is considered. Otherwise the whole document is.
	subtree := doc.RawNode
	if len(inputDoc.Pointer) > 0 {
		if subtree = doc.GetByPath(inputDoc.Pointer); subtree == nil {
			return fmt.Errorf("%w: node %q not found in the document", common2.ErrInvalidCLIArgument, inputDoc.PointerString())
		}
	}
	logger.Debug("Using target subtree", "pointer", jsonpointer.PointerString(subtree.Path()...))

	// Collecting nodes to generate examples for.
	useSchemas, useMessages := cmdConfig.Doc.GenExamples.OnlySchemas, cmdConfig.Doc.GenExamples.OnlyMessages
	if !useSchemas && !useMessages {
		useSchemas, useMessages = true, true // Select all these entities by default
	}
	logger.Debug("Collecting target nodes", "useSchemas", useSchemas, "useMessages", useMessages)
	targets := exampleCollectTargets(subtree, nil, useSchemas, useMessages)
	if len(targets) == 0 {
		logger.Warn("No matching nodes found, nothing to do")
		return nil
	}

	originDoc := doc.AbsOriginDocumentPath()
	var generated int
	docs := map[string]*documentTree{originDoc.String(): doc}
	for _, t := range targets {
		examplesNode, hasExample := t.Get("examples")
		if !hasExample {
			examplesNode = types.NewEmptyRawNode(types.RawNodeKindArray, append(t.Path(), "examples"), originDoc)
		}
		if examplesNode.Kind() != types.RawNodeKindArray {
			logger.Warn("Existing examples subnode is not an array, skipping", "path", t.AbsPointerString(), "kind", examplesNode.Kind())
			break
		}
		if (hasExample && examplesNode.Len() > 0) && !cmdConfig.Doc.GenExamples.Append {
			logger.Debug("Skipping a node because it already contains the examples", "path", t.AbsPointerString())
			continue
		}

		entity := asyncapiEntityByPath(t.Path())

		logger.Debug("Generating examples", "path", t.AbsPointerString(), "entity", entity, "count", cmdConfig.Doc.GenExamples.Count, "alreadyHasExample", hasExample)
		for range cmdConfig.Doc.GenExamples.Count {
			// Generate an example for the node. Note that the node and all nested nodes have empty path
			example, err := exampleGenerateNode(entity, t, docs, originDoc, locator, cmdConfig)
			if err != nil {
				logger.Error("Error occurred while generating example for node, skipping", "path", t.AbsPointerString(), "error", err)
				continue
			}
			logger.Trace("Generated example", "path", t.Path(), "example", example)
			if example == nil {
				logger.Debug("Example generation returned nil, skipping", "path", t.AbsPointerString())
				continue // Example has not been generated by some reason
			}
			examplesNode.Append(example)
			generated++
		}
		t.Set("examples", examplesNode)
	}

	logger.Info("Writing document", "file", outputPath, "examplesGenerated", generated)
	buf := bytes.NewBuffer(nil)
	enc, err := getDocumentEncoder(buf, cmdConfig)
	if err != nil {
		return fmt.Errorf("get encoder: %w", err)
	}
	if err = enc.Encode(doc); err != nil {
		return fmt.Errorf("marshal document: %w", err)
	}
	if err = os.WriteFile(outputPath, buf.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write output file %q: %w", outputPath, err)
	}

	return nil
}

// exampleCollectTargets recursively walks the node and returns every nested node that has one of the selected entity types.
func exampleCollectTargets(node *types.RawNode, visitedTypes []string, schemas, messages bool) []*types.RawNode {
	logger := log.GetLogger("")

	if node == nil || node.Kind() == types.RawNodeKindScalar {
		logger.Trace("Skipping node because it is nil or scalar")
		return nil
	}

	entity := asyncapiEntityByPath(node.Path())
	logger.Trace("Visiting node", "path", node.Path(), "entity", entity)

	// Consider only the first occurrence of a node with a given entity in a tree branch.
	// This way we avoid to generate examples recursively for jsonschemas, but able do this for the
	// jsonschemas that could met in entities of other entitites, e.g. ones in message bindings.
	if slices.Contains(visitedTypes, entity) {
		logger.Trace("Skipping node because its entity has already been visited in this branch", "path", node.Path(), "entity", entity, "visitedTypes", visitedTypes)
		return nil
	}
	if strings.HasPrefix(entity, ">") {
		visitedTypes = append(visitedTypes, entity)
	}

	matchEntity := entity == ">message" && messages || entity == ">schema" && schemas
	var res []*types.RawNode
	if matchEntity && !node.Has("$ref") {
		logger.Trace("Matched node", "path", node.Path(), "entity", entity, "visitedTypes", visitedTypes)
		res = append(res, node)
	}

	logger.Trace("Iterating node entries", "path", node.Path(), "entity", entity, "visitedTypes", visitedTypes)
	for _, e := range node.Entries() {
		res = append(res, exampleCollectTargets(e, visitedTypes, schemas, messages)...)
	}
	return res
}

func exampleGenerateNode(entity string, node *types.RawNode, docs map[string]*documentTree, originDoc *jsonpointer.JSONPointer, locator common2.DocumentLocator, cmdConfig common2.ToolConfig) (*types.RawNode, error) {
	res := types.NewEmptyRawNode(types.RawNodeKindObject, nil, originDoc)

	switch entity {
	case ">schema":
		return exampleGenerateSchema(node, docs, nil, originDoc, locator, cmdConfig)
	case ">message":
		if payload, ok := node.Get("payload"); ok {
			n, err := exampleGenerateSchema(payload, docs, nil, originDoc, locator, cmdConfig)
			if err != nil {
				return nil, fmt.Errorf("generate payload example: %w", err)
			}
			if n != nil {
				res.Set("payload", n)
			}
		}
		fallthrough
	case ">messageTrait":
		if headers, ok := node.Get("headers"); ok {
			n, err := exampleGenerateSchema(headers, docs, nil, originDoc, locator, cmdConfig)
			if err != nil {
				return nil, fmt.Errorf("generate headers example: %w", err)
			}
			if n != nil {
				res.Set("headers", n)
			}
		}
		if res.IsZero() {
			return nil, nil // Neither payload nor headers schema present, nothing to generate
		}

		messageName := lo.PascalCase(strings.Join(node.Path(), "_"))
		if n, ok := node.Get("name"); ok && n.AsStringSafe() != "" {
			messageName = n.AsStringSafe()
		} else if n, ok = node.Get("title"); ok && n.AsStringSafe() != "" {
			messageName = n.AsStringSafe()
		}
		res.SetAtStart("name", types.NewScalarRawNode(nil, fmt.Sprintf("Example for message %s", messageName), originDoc))
	}

	return res, nil
}

func exampleGenerateSchema(node *types.RawNode, docs map[string]*documentTree, visited []*types.RawNode, originDoc *jsonpointer.JSONPointer, locator common2.DocumentLocator, cmdConfig common2.ToolConfig) (*types.RawNode, error) {
	if node == nil || node.Kind() != types.RawNodeKindObject {
		return types.NewScalarRawNode(nil, nil, originDoc), nil
	}

	// Follow a $ref to the schema it points to. Only local refs are resolved; external/remote refs are left as a
	// null value, since example generation does not load other documents.
	if node.Has("$ref") {
		visited = append(visited, node)

		ref, err := parseRefRawNode(node)
		if err != nil {
			return nil, fmt.Errorf("parse $ref: %w", err)
		}
		if !cmdConfig.Doc.GenExamples.AllowRemoteReferences && ref.URI != nil {
			return nil, fmt.Errorf(
				"%s: external requests are forbidden by default for security reasons, use --allow-remote-refs flag to allow them",
				ref,
			)
		}
		target, err := resolveRefNode(ref, node, docs, locator, true, true)
		if err != nil {
			return nil, fmt.Errorf("resolve $ref: %w", err)
		}
		if slices.Contains(visited, target) {
			// TODO: check required, return error
			return nil, nil // Stop if we have already visited this target in recursive schema
		}

		return exampleGenerateSchema(target, docs, visited, originDoc, locator, cmdConfig)
	}

	objectTypes, err := getSchemaType(node)
	if err != nil {
		return nil, fmt.Errorf("get schema type at %q: %w", node.AbsPointerString(), err)
	}
	for _, typ := range objectTypes {
		switch typ {
		case "object":
			var key string
			switch {
			case node.Has("allOf"):
				// allOf is a conjunction of subschemas; merge the objects generated from each of them.
				return exampleMergeAllOf(node, docs, visited, originDoc, locator, cmdConfig)
			case node.Has("oneOf"):
				key = "oneOf"
				fallthrough
			case node.Has("anyOf"):
				if key == "" {
					key = "anyOf"
				}
				n, _ := node.Get(key)
				if n.Kind() != types.RawNodeKindArray {
					return nil, fmt.Errorf("%q is not an array at %q", key, node.AbsPointerString())
				}

				// anyOf/oneOf is a disjunction; the first subschema is enough for an example.
				var item *types.RawNode
				for _, e := range n.Entries() {
					if e.Kind() != types.RawNodeKindObject {
						return nil, fmt.Errorf("item in %q is not an object at %q", key, e.AbsPointerString())
					}
					item = e
					break
				}

				visited = append(visited, item)

				return exampleGenerateSchema(item, docs, visited, originDoc, locator, cmdConfig)
			}
			return exampleGenerateObject(node, docs, visited, originDoc, locator, cmdConfig)
		case "array":
			return exampleGenerateArray(node, docs, visited, originDoc, locator, cmdConfig)
		default:
			return types.NewScalarRawNode(nil, exampleGenerateScalar(node, typ, cmdConfig), originDoc), nil
		}
	}

	return nil, fmt.Errorf("no supported type found at %q", node.AbsPointerString())
}

func getSchemaType(node *types.RawNode) ([]string, error) {
	value, ok := node.Get("type")
	if !ok {
		// Guess the object type euristically if the type was not specified, which is allowed in some jsonschema specs.
		switch {
		case !node.Has("$ref") && node.Has("properties"):
			return []string{"object"}, nil
		case node.Has("items"): // TODO: fix type when AllOf, AnyOf, OneOf
			return []string{"array"}, nil
		default:
			return []string{"object"}, nil
		}
	}

	switch value.Kind() {
	case types.RawNodeKindScalar:
		s, ok := value.AsScalar().(string)
		if !ok {
			return nil, fmt.Errorf("not a string")
		}
		return []string{s}, nil
	case types.RawNodeKindArray:
		var res []string
		for _, e := range value.Entries() {
			if e.Kind() != types.RawNodeKindScalar {
				return nil, fmt.Errorf("not a scalar value")
			}
			s, ok := e.AsScalar().(string)
			if !ok {
				return nil, fmt.Errorf("not a string")
			}
			res = append(res, s)
		}
		return res, nil
	case types.RawNodeKindObject:
		return nil, fmt.Errorf("not a string or an array")
	}

	panic(fmt.Errorf("invalid node kind %q", value.Kind()))
}

func exampleMergeAllOf(node *types.RawNode, docs map[string]*documentTree, visited []*types.RawNode, originDoc *jsonpointer.JSONPointer, locator common2.DocumentLocator, cmdConfig common2.ToolConfig) (*types.RawNode, error) {
	allOfNode, ok := node.Get("allOf")
	if !ok || allOfNode.Kind() != types.RawNodeKindArray {
		return nil, fmt.Errorf("allOf is not an array at %q", node.AbsPointerString())
	}

	visited = append(visited, allOfNode)

	res := types.NewEmptyRawNode(types.RawNodeKindObject, nil, originDoc)
	for _, e := range allOfNode.Entries() {
		part, err := exampleGenerateSchema(e, docs, visited, originDoc, locator, cmdConfig)
		if err != nil {
			return nil, err
		}
		if part == nil || part.Kind() != types.RawNodeKindObject {
			continue
		}
		for key, val := range part.Entries() {
			s, ok := key.(string)
			if !ok {
				return nil, fmt.Errorf("non-string key in object at %q", part.AbsPointerString())
			}
			res.Set(s, val)
		}
	}
	return res, nil
}

func exampleGenerateObject(node *types.RawNode, docs map[string]*documentTree, visited []*types.RawNode, originDoc *jsonpointer.JSONPointer, locator common2.DocumentLocator, cmdConfig common2.ToolConfig) (*types.RawNode, error) {
	propsNode, ok := node.Get("properties")
	if !ok || propsNode.Kind() != types.RawNodeKindObject {
		return nil, fmt.Errorf("properties node is not an object at %q", node.AbsPointerString())
	}

	res := types.NewEmptyRawNode(types.RawNodeKindObject, nil, originDoc)
	visited = append(visited, propsNode)

	for key, propSchema := range propsNode.Entries() {
		v, err := exampleGenerateSchema(propSchema, docs, visited, originDoc, locator, cmdConfig)
		if err != nil {
			return nil, err
		}
		if v != nil {
			res.Set(key, v)
		}
	}

	return res, nil
}

func exampleGenerateArray(node *types.RawNode, docs map[string]*documentTree, visited []*types.RawNode, originDoc *jsonpointer.JSONPointer, locator common2.DocumentLocator, cmdConfig common2.ToolConfig) (*types.RawNode, error) {
	itemsNode, ok := node.Get("items")
	if !ok || itemsNode.Kind() != types.RawNodeKindObject {
		return nil, fmt.Errorf("items node is not an object at %q", node.AbsPointerString())
	}
	res := types.NewEmptyRawNode(types.RawNodeKindArray, nil, originDoc)
	visited = append(visited, itemsNode)

	// A single item is enough to illustrate the array shape.
	v, err := exampleGenerateSchema(itemsNode, docs, visited, originDoc, locator, cmdConfig)
	if err != nil {
		return nil, err
	}
	if v != nil {
		res.Append(v)
	}

	return res, nil
}

func exampleGenerateScalar(node *types.RawNode, typ string, cmdConfig common2.ToolConfig) any {
	var format string
	if formatNode, ok := node.Get("format"); ok {
		format = formatNode.AsStringSafe()
	}

	switch typ {
	case "boolean":
		return gofakeit.Bool()
	case "integer":
		return gofakeit.Number(0, 1000)
	case "number":
		return gofakeit.Float64Range(0, 1000)
	case "null":
		return nil
	case "string":
		return exampleGenerateString(format, cmdConfig)
	default:
		return exampleGenerateString(format, cmdConfig)
	}
}

func exampleGenerateString(format string, cmdConfig common2.ToolConfig) string {
	switch format {
	case "date-time", "datetime":
		return gofakeit.Date().Format(cmdConfig.Doc.GenExamples.DateTimeFormat)
	case "date":
		return gofakeit.Date().Format(cmdConfig.Doc.GenExamples.DateFormat)
	case "time":
		return gofakeit.Date().Format(cmdConfig.Doc.GenExamples.TimeFormat)
	case "email", "idn-email":
		return gofakeit.Email()
	case "uuid":
		return gofakeit.UUID()
	case "uri", "url", "uri-reference", "iri":
		return gofakeit.URL()
	case "hostname", "idn-hostname":
		return gofakeit.DomainName()
	case "ipv4":
		return gofakeit.IPv4Address()
	case "ipv6":
		return gofakeit.IPv6Address()
	case "password":
		return gofakeit.Password(true, true, true, false, false, 12)
	default:
		return gofakeit.Word()
	}
}
