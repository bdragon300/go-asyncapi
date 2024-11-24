package http

type ProtoBuilder struct {
	ProtoName string
}

var Builder = ProtoBuilder{
	ProtoName:  "http",
}

func (pb ProtoBuilder) ProtocolName() string {
	return pb.ProtoName
}
