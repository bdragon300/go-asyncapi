package amqp

type ProtoBuilder struct {
	ProtoName, ProtoTitle string
}

var Builder = ProtoBuilder{
	ProtoName:  "amqp",
	ProtoTitle: "AMQP",
}

func (pb ProtoBuilder) ProtocolName() string {
	return pb.ProtoName
}

func (pb ProtoBuilder) ProtocolTitle() string {
	return pb.ProtoTitle
}