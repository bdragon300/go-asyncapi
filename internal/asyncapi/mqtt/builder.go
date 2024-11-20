package mqtt

type ProtoBuilder struct {
	ProtoName, ProtoTitle string
}

var Builder = ProtoBuilder{
	ProtoName:  "mqtt",
	ProtoTitle: "MQTT",
}

func (pb ProtoBuilder) ProtocolName() string {
	return pb.ProtoName
}

func (pb ProtoBuilder) ProtocolTitle() string {
	return pb.ProtoTitle
}