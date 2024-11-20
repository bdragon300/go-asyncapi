package ip

type ProtoBuilder struct {
	ProtoName, ProtoTitle string
}

var Builder = ProtoBuilder{
	ProtoName:  "ip",
	ProtoTitle: "IP",
}

func (pb ProtoBuilder) ProtocolName() string {
	return pb.ProtoName
}

func (pb ProtoBuilder) ProtocolTitle() string {
	return pb.ProtoTitle
}