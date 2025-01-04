package franzgo

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	runKafka "github.com/bdragon300/go-asyncapi/run/kafka"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kversion"
)

func NewProducer(serverURL string, bindings *runKafka.ServerBindings, extraOpts []kgo.Opt) (*ProduceClient, error) {
	return &ProduceClient{
		serverURL: serverURL,
		bindings:  bindings,
		extraOpts: extraOpts,
	}, nil
}

type ProduceClient struct {
	serverURL string
	bindings  *runKafka.ServerBindings
	extraOpts []kgo.Opt
}

func (p ProduceClient) Publisher(_ context.Context, address string, bindings *runKafka.ChannelBindings) (runKafka.Publisher, error) {
	// TODO: schema registry https://github.com/twmb/franz-go/blob/master/examples/schema_registry/schema_registry.go
	var opts []kgo.Opt

	u, err := url.Parse(p.serverURL)
	if err != nil {
		return nil, fmt.Errorf("url parse: %w", err)
	}
	opts = append(opts, kgo.SeedBrokers(strings.Split(u.Host, ",")...))

	topic := address
	if bindings != nil && bindings.Topic != "" {
		topic = bindings.Topic
	}
	if topic != "" {
		opts = append(opts, kgo.DefaultProduceTopic(topic))
	}
	opts = append(opts, p.extraOpts...)

	cl, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	return &PublishChannel{
		Client:   cl,
		Topic:    topic,
		bindings: bindings,
	}, nil
}

type ImplementationRecord interface {
	AsFranzGoRecord() *kgo.Record
	// TODO: Bindings?
}

type PublishChannel struct {
	*kgo.Client
	Topic    string
	bindings *runKafka.ChannelBindings
}

func (p PublishChannel) Send(ctx context.Context, envelopes ...runKafka.EnvelopeWriter) error {
	records := make([]*kgo.Record, 0, len(envelopes))
	for _, e := range envelopes {
		rm := e.(ImplementationRecord)
		records = append(records, rm.AsFranzGoRecord())
	}
	return p.Client.ProduceSync(ctx, records...).FirstErr()
}

func (p PublishChannel) Close() error {
	p.Client.Close()
	return nil
}

func ParseProtocolVersion(protocolVersion string) (*kversion.Versions, error) {
	var ver *kversion.Versions
	switch protocolVersion {
	case "stable":
		ver = kversion.Stable()
	case "tip":
		ver = kversion.Tip()
	case "0.8.0":
		ver = kversion.V0_8_0()
	case "0.8.1":
		ver = kversion.V0_8_1()
	case "0.8.2":
		ver = kversion.V0_8_2()
	case "0.9.0":
		ver = kversion.V0_9_0()
	case "0.10.0":
		ver = kversion.V0_10_0()
	case "0.10.1":
		ver = kversion.V0_10_1()
	case "0.10.2":
		ver = kversion.V0_10_2()
	case "0.11.0":
		ver = kversion.V0_11_0()
	case "1.0.0":
		ver = kversion.V1_0_0()
	case "1.1.0":
		ver = kversion.V1_1_0()
	case "2.0.0":
		ver = kversion.V2_0_0()
	case "2.1.0":
		ver = kversion.V2_1_0()
	case "2.2.0":
		ver = kversion.V2_2_0()
	case "2.3.0":
		ver = kversion.V2_3_0()
	case "2.4.0":
		ver = kversion.V2_4_0()
	case "2.5.0":
		ver = kversion.V2_5_0()
	case "2.6.0":
		ver = kversion.V2_6_0()
	case "2.7.0":
		ver = kversion.V2_7_0()
	case "2.8.0":
		ver = kversion.V2_8_0()
	case "3.0.0":
		ver = kversion.V3_0_0()
	case "3.1.0":
		ver = kversion.V3_1_0()
	case "3.2.0":
		ver = kversion.V3_2_0()
	case "3.3.0":
		ver = kversion.V3_3_0()
	case "3.4.0":
		ver = kversion.V3_4_0()
	case "3.5.0":
		ver = kversion.V3_5_0()
	default:
		return nil, fmt.Errorf("unknown protocol version: %s", protocolVersion)
	}

	return ver, nil
}
