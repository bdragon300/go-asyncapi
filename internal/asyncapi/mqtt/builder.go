package mqtt

type ProtoBuilder struct {
	ProtoName string
}

var Builder = ProtoBuilder{
	ProtoName: "mqtt",
}

func (pb ProtoBuilder) ProtocolName() string {
	return pb.ProtoName
}
