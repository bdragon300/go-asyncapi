package udp

type ProtoBuilder struct {
	ProtoName string
}

var Builder = ProtoBuilder{
	ProtoName:  "udp",
}

func (pb ProtoBuilder) ProtocolName() string {
	return pb.ProtoName
}
