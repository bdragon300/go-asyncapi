package common

type PackageKind string

const (
	RuntimePackageKind  PackageKind = "runtime"
	ModelsPackageKind   PackageKind = "models"
	MessagesPackageKind PackageKind = "messages"
	ChannelsPackageKind PackageKind = "channels"
	ServersPackageKind  PackageKind = "servers"
)
