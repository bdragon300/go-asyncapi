package common

type PackageKind string

const (
	RuntimePackageKind  PackageKind = "runtime"
	ModelsPackageKind   PackageKind = "models"
	MessagesPackageKind PackageKind = "messages"
	ChannelsPackageKind PackageKind = "channels"
	ServersPackageKind  PackageKind = "servers"
)

type SchemaTag string

const (
	SchemaTagNoInline    SchemaTag = "noinline"
	SchemaTagPackageDown SchemaTag = "packageDown"
)

const TagName = "cgen"

