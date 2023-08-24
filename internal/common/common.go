package common

type PackageKind string

const (
	DefaultPackageKind PackageKind = "default"
	ModelsPackageKind  PackageKind = "models"
	MessagePackageKind PackageKind = "message"
)
