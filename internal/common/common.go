package common

type PackageKind string

const (
	ModelsPackageKind     PackageKind = "models"
	MessagesPackageKind   PackageKind = "messages"
	ChannelsPackageKind   PackageKind = "channels"
	ServersPackageKind    PackageKind = "servers"
	ParametersPackageKind PackageKind = "parameters"

	RuntimePackageKind      PackageKind = "runtime"
	RuntimeKafkaPackageKind PackageKind = "runtime/kafka"
	RuntimeAMQPPackageKind  PackageKind = "runtime/amqp"
)

type SchemaTag string

const (
	SchemaTagNoInline    SchemaTag = "noinline"
	SchemaTagPackageDown SchemaTag = "packageDown"
)

const TagName = "cgen"

