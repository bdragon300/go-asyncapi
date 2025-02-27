package common

// SchemaTagName is a tag that are used to set quirks for [asyncapi] entities
const SchemaTagName = "cgen"

type SchemaTag string

const (
	// SchemaTagSelectable marks the objects available be selected from tool config. Typically, this tag is used for
	// entities that are defined in the sections `channels`, `servers`, etc.
	SchemaTagSelectable SchemaTag = "selectable"

	// SchemaTagDefinition forces an object to be rendered as a Go definition instead of inlined object.
	//
	// For example, the nested jsonschemas defined in $allOf, $anyOf, $oneOf sections should be rendered as separate
	// structs that are used in union struct.
	SchemaTagDefinition SchemaTag = "definition"

	// SchemaTagComponent is special tag for top-level entities in `components` section. Basically it affects if the
	// jsonschema object are considered as schema with [ArtifactKindSchema] kind or inlined object with [ArtifactKindOther] kind.
	SchemaTagComponent SchemaTag = "components"

	// SchemaTagDataModel marks the jsonschema objects and all its nested objects as data models.
	// In particular, the tags in fields in data models are driven by document and message content types.
	//
	// Typically, this tag marks the message payload and header entities and also common-used document schemas.
	SchemaTagDataModel SchemaTag = "data_model"
)
