package kafka

type ProtoBuilder struct {
	ProtoName, ProtoTitle string
}

var Builder = ProtoBuilder{
	ProtoName:  "kafka",
	ProtoTitle: "Kafka",
}

func (pb ProtoBuilder) ProtocolName() string {
	return pb.ProtoName
}

func (pb ProtoBuilder) ProtocolTitle() string {
	return pb.ProtoTitle
}