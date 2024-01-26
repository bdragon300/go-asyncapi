package common

type SchemaTag string

const (
	SchemaTagDirectRender SchemaTag = "directRender" // Object must be rendered directly (only on the current level)
	SchemaTagPkgScope     SchemaTag = "pkgScope"
	SchemaTagComponent    SchemaTag = "components" // Set to objects located in `components` document section
)
