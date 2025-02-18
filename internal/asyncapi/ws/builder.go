package ws

type ProtoBuilder struct {
	ProtoName string
}

var Builder = ProtoBuilder{
	ProtoName: "ws",
}

func (pb ProtoBuilder) ProtocolName() string {
	return pb.ProtoName
}
