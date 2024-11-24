package kafka

type ProtoBuilder struct {
	ProtoName string
}

var Builder = ProtoBuilder{
	ProtoName:  "kafka",
}

func (pb ProtoBuilder) ProtocolName() string {
	return pb.ProtoName
}
