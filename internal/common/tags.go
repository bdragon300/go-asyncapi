package common

type SchemaTag string

const (
	// SchemaTagDefinition forces an object to be rendered as a definition instead of inline declaration
	SchemaTagDefinition SchemaTag = "definition"  // TODO: what the difference between this and SchemaTagComponent?
	// SchemaTagComponent marks all top-level objects in `component` section
	SchemaTagComponent SchemaTag = "components"
	// SchemaTagMarshal marks that an object is meant to be marshaled/unmarshaled. Inherited to the nested objects
	SchemaTagMarshal SchemaTag = "marshal"
)
