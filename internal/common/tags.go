package common

type SchemaTag string

const (
	// SchemaTagDirectRender marks that an object must have directRender=true
	SchemaTagDirectRender SchemaTag = "directRender"
	// SchemaTagPkgScope sets the package scope for an object. Inherited by nested objects
	SchemaTagPkgScope SchemaTag = "pkgScope"
	// SchemaTagComponent marks all top-level objects in `component` section
	SchemaTagComponent SchemaTag = "components"
	// SchemaTagMarshal marks that an object is meant to be marshaled/unmarshaled. Inherited by nested objects
	SchemaTagMarshal SchemaTag = "marshal"
)
