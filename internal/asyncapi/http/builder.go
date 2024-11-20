package http

type ProtoBuilder struct {
	ProtoName, ProtoTitle string
}

var Builder = ProtoBuilder{
	ProtoName:  "http",
	ProtoTitle: "HTTP",
}

func (pb ProtoBuilder) ProtocolName() string {
	return pb.ProtoName
}

func (pb ProtoBuilder) ProtocolTitle() string {
	return pb.ProtoTitle
}