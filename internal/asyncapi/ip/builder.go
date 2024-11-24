package ip

type ProtoBuilder struct {
	ProtoName string
}

var Builder = ProtoBuilder{
	ProtoName:  "ip",
}

func (pb ProtoBuilder) ProtocolName() string {
	return pb.ProtoName
}
