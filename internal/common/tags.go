package common

// SchemaTagName is a tag that are used to set quirks for [asyncapi] entities
const SchemaTagName = "cgen"

type SchemaTag string

const (
	// SchemaTagSelectable marks the objects that should get to selections, i.e. basically objects that are
	// rendered directly by feeding to the root template. See [internal/renderer/code.go RenderArtifacts].
	SchemaTagSelectable SchemaTag = "selectable"

	// SchemaTagDataModel marks the jsonschema objects and all its nested objects as data models.
	// In particular, the tags in fields in data models are driven by document and message content types.
	//
	// Typically, this tag marks the message payload and header entities and also common-used document schemas.
	SchemaTagDataModel SchemaTag = "data_model"
)
