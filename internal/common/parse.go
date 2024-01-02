package common

const TagName = "cgen"

type SchemaTag string

const (
	SchemaTagDirectRender SchemaTag = "directRender"
	SchemaTagPackageDown  SchemaTag = "packageDown"
)

// DefaultPackage is package where objects will put if current parse package is empty
const DefaultPackage = "default"
