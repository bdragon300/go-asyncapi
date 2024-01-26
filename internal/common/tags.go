package common

type SchemaTag string

const (
	SchemaTagDirectRender SchemaTag = "directRender"
	SchemaTagPkgScope     SchemaTag = "pkgScope"
	SchemaTagCompoennt    SchemaTag = "components" // Set to objects located in `components` document section
)
