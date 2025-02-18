package redis

type ProtoBuilder struct {
	ProtoName string
}

var Builder = ProtoBuilder{
	ProtoName: "redis",
}

func (pb ProtoBuilder) ProtocolName() string {
	return pb.ProtoName
}
