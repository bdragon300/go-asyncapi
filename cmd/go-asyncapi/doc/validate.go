package doc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/bdragon300/go-asyncapi/assets"
	common2 "github.com/bdragon300/go-asyncapi/cmd/go-asyncapi/common"
	"github.com/bdragon300/go-asyncapi/internal/compiler"
	"github.com/bdragon300/go-asyncapi/internal/jsonpointer"
	"github.com/bdragon300/go-asyncapi/internal/log"
	"github.com/santhosh-tekuri/jsonschema/v6"
	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// builtInSchemaDir is the directory inside AssetFS that holds the AsyncAPI JSON Schema files
const builtInSchemaDir = "schemas"

// validationMessagePrinter renders the localized text of a single schema violation.
var validationMessagePrinter = message.NewPrinter(language.English)

type ValidateCmd struct {
	Documents []string `arg:"positional,required" help:"AsyncAPI document file or URL" placeholder:"FILE"`
	Schema    string   `arg:"--schema,-s" help:"Custom JSON Schema file to validate against. By default, the built-in AsyncAPI schema matching the document version is used." placeholder:"FILE"`
}

func cliValidate(cmd *ValidateCmd, cmdConfig common2.ToolConfig) error {
	logger := log.GetLogger("")
	locator := common2.GetLocator(cmdConfig)

	// Optional custom schema, compiled once and reused for every document.
	var customSchema *jsonschema.Schema
	if cmdConfig.Doc.Validate.Schema != "" {
		f, err := os.Open(cmdConfig.Doc.Validate.Schema)
		if err != nil {
			return fmt.Errorf("open custom schema %q: %w", cmdConfig.Doc.Validate.Schema, err)
		}
		s, err := compileSchemaFile(path.Base(cmdConfig.Doc.Validate.Schema), f)
		_ = f.Close()
		if err != nil {
			return fmt.Errorf("load custom schema %q: %w", cmdConfig.Doc.Validate.Schema, err)
		}
		logger.Debug("Using custom schema", "file", cmdConfig.Doc.Validate.Schema)
		customSchema = s
	}

	var invalidCount int
	for _, doc := range cmd.Documents {
		docURL, err := jsonpointer.Parse(doc)
		if err != nil {
			return fmt.Errorf("parse document url: %w", err)
		}
		valid, err := validateDocument(docURL, customSchema, locator)
		if err != nil {
			return fmt.Errorf("validate document %q: %w", doc, err)
		}
		if !valid {
			invalidCount++
		}
	}

	if invalidCount > 0 {
		return fmt.Errorf("%w: %d of %d document(s) failed validation", common2.ErrBadResult, invalidCount, len(cmd.Documents))
	}
	logger.Info("All documents are valid", "count", len(cmd.Documents))
	return nil
}

// validateDocument reads, normalizes and validates a single document. It logs every
// validation error found and returns whether the document is valid. A returned error
// means the document could not be processed at all (e.g. read/parse failure).
func validateDocument(docURL *jsonpointer.JSONPointer, customSchema *jsonschema.Schema, locator common2.DocumentLocator) (bool, error) {
	logger := log.GetLogger("")

	buf, newDecoder, err := compiler.ReadDocument(docURL, locator, logger)
	if err != nil {
		return false, fmt.Errorf("read: %w", err)
	}

	// Decode the document (YAML or JSON), then normalize it into a JSON instance so
	// that both formats are validated uniformly by the JSON Schema validator.
	var raw any
	if err = newDecoder(bytes.NewReader(buf)).Decode(&raw); err != nil {
		return false, fmt.Errorf("decode: %w", err)
	}
	jsonBuf, err := json.Marshal(raw)
	if err != nil {
		return false, fmt.Errorf("normalize document to JSON: %w", err)
	}
	instance, err := jsonschema.UnmarshalJSON(bytes.NewReader(jsonBuf))
	if err != nil {
		return false, fmt.Errorf("parse document: %w", err)
	}

	schema := customSchema
	if schema == nil {
		version := documentVersion(raw)
		if version == "" {
			return false, fmt.Errorf("document does not specify an \"asyncapi\" version; use --schema to provide a schema explicitly")
		}
		if schema, err = loadBuiltInSchema(version); err != nil {
			return false, err
		}
		logger.Debug("Using built-in AsyncAPI schema", "version", version)
	}

	logger.Info("Validating document", "document", docURL)
	if err = schema.Validate(instance); err != nil {
		var ve *jsonschema.ValidationError
		if !errors.As(err, &ve) {
			logger.Error("Validation error", "document", docURL, "error", err.Error())
			return false, nil
		}
		leaves := leafValidationErrors(ve)
		for _, e := range leaves {
			logger.Error(
				"Validation error",
				"document", docURL,
				"location", jsonpointer.PointerString(e.InstanceLocation...),
				"error", e.ErrorKind.LocalizedString(validationMessagePrinter),
			)
		}
		logger.Error("Document is invalid", "document", docURL, "errors", len(leaves))
		return false, nil
	}

	logger.Info("Document is valid", "document", docURL)
	return true, nil
}

// documentVersion extracts the value of the top-level "asyncapi" field, or "" if absent.
func documentVersion(raw any) string {
	m, ok := raw.(map[string]any)
	if !ok {
		return ""
	}
	v, _ := m["asyncapi"].(string)
	return v
}

func loadBuiltInSchema(version string) (*jsonschema.Schema, error) {
	name := builtInSchemaDir + "/" + version + ".json"
	f, err := assets.AssetFS.Open(name)
	if err != nil {
		return nil, fmt.Errorf(
			"AsyncAPI version %q is not supported by default; use --schema to provide a custom schema file", version,
		)
	}
	defer f.Close()
	s, err := compileSchemaFile(name, f)
	if err != nil {
		return nil, fmt.Errorf("compile embedded schema for version %q: %w", version, err)
	}

	return s, nil
}

// compileSchemaFile reads a JSON Schema from r and compiles it. If the schema declares a
// root "$id", that identifier is used as its base URI so internal references resolve.
func compileSchemaFile(loc string, r io.Reader) (*jsonschema.Schema, error) {
	schemaDoc, err := jsonschema.UnmarshalJSON(r)
	if err != nil {
		return nil, fmt.Errorf("parse schema: %w", err)
	}

	if m, ok := schemaDoc.(map[string]any); ok {
		if id, ok := m["$id"].(string); ok && id != "" {
			loc = id
		}
	}

	c := jsonschema.NewCompiler()
	if err = c.AddResource(loc, schemaDoc); err != nil {
		return nil, fmt.Errorf("add schema resource: %w", err)
	}
	sch, err := c.Compile(loc)
	if err != nil {
		return nil, fmt.Errorf("compile schema: %w", err)
	}
	return sch, nil
}

// leafValidationErrors flattens the validation error tree into its leaves. The leaf
// nodes carry the most specific messages (e.g. "missing property 'version'"), whereas
// inner nodes only describe the failing combinator (allOf/oneOf/...).
func leafValidationErrors(ve *jsonschema.ValidationError) []*jsonschema.ValidationError {
	if len(ve.Causes) == 0 {
		return []*jsonschema.ValidationError{ve}
	}
	var res []*jsonschema.ValidationError
	for _, c := range ve.Causes {
		res = append(res, leafValidationErrors(c)...)
	}
	return res
}
