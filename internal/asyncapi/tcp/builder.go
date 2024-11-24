package tcp

type ProtoBuilder struct {
	ProtoName string
}

var Builder = ProtoBuilder{
	ProtoName:  "tcp",
}

func (pb ProtoBuilder) ProtocolName() string {
	return pb.ProtoName
}
