package common

type SchemaTag string

const (
	// SchemaTagSelectable force marks that an object is selectable. This makes the object to be rendered if it is
	// not selectable, e.g. it's defined in `components` section
	SchemaTagSelectable SchemaTag = "selectable" // SchemaTagSelectable marks that an object is selectable
	// SchemaTagDefinition forces an object to be rendered as a definition instead of inline declaration
	SchemaTagDefinition SchemaTag = "definition"
	// SchemaTagComponent marks all top-level objects in `component` section
	SchemaTagComponent SchemaTag = "components"
	// SchemaTagMarshal marks that an object is meant to be marshaled/unmarshaled. Inherited to the nested objects
	SchemaTagMarshal SchemaTag = "marshal"
)
