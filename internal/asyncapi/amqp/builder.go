package amqp

type ProtoBuilder struct {
	ProtoName string
}

var Builder = ProtoBuilder{
	ProtoName: "amqp",
}

func (pb ProtoBuilder) ProtocolName() string {
	return pb.ProtoName
}
