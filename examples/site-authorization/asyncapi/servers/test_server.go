// Code generated by go-asyncapi tool. DO NOT EDIT.

package servers

import (
	"context"
	"errors"
	"github.com/bdragon300/go-asyncapi/run/kafka"
	"github.com/twmb/franz-go/pkg/kgo"
	"io"
	"net/url"
	"site-authorization/asyncapi/channels"
	kafka2 "site-authorization/asyncapi/impl/kafka"
	"site-authorization/asyncapi/operations"
)

func TestServerURL() (*url.URL, error) {
	return &url.URL{Scheme: "kafka", Host: "example.com:9092", Path: ""}, nil
}

func NewTestServer(producer kafka.Producer, consumer kafka.Consumer) *TestServer {
	return &TestServer{
		producer: producer,
		consumer: consumer,
	}
}

type TestServerClosable struct {
	TestServer
}

func (c TestServerClosable) Close() error {
	var err error
	if v, ok := any(c.producer).(io.Closer); ok {
		err = errors.Join(err, v.Close())
	}
	if v, ok := any(c.consumer).(io.Closer); ok {
		err = errors.Join(err, v.Close())
	}
	return err
}

func ConnectTestServerBidi(ctx context.Context, url *url.URL, opts ...kgo.Opt) (*TestServerClosable, error) {
	var bindings *kafka.ServerBindings
	producer := kafka2.NewProducer([]string{url.Host}, bindings, opts...)
	consumer := kafka2.NewConsumer([]string{url.Host}, bindings, opts...)
	return &TestServerClosable{
		TestServer{producer: producer, consumer: consumer},
	}, nil
}

func ConnectTestServerProducer(ctx context.Context, url *url.URL, opts ...kgo.Opt) (*TestServerClosable, error) {
	var bindings *kafka.ServerBindings
	producer := kafka2.NewProducer([]string{url.Host}, bindings, opts...)
	return &TestServerClosable{
		TestServer{producer: producer},
	}, nil
}

func ConnectTestServerConsumer(ctx context.Context, url *url.URL, opts ...kgo.Opt) (*TestServerClosable, error) {
	var bindings *kafka.ServerBindings
	consumer := kafka2.NewConsumer([]string{url.Host}, bindings, opts...)
	return &TestServerClosable{
		TestServer{consumer: consumer},
	}, nil
}

// TestServer--Main Kafka broker
type TestServer struct {
	producer kafka.Producer
	consumer kafka.Consumer
}

func (s TestServer) Name() string {
	return "TestServer"
}

func (s TestServer) Producer() kafka.Producer {
	return s.producer
}

func (s TestServer) Consumer() kafka.Consumer {
	return s.consumer
}

func (s TestServer) OpenAuthChannelKafka(
	ctx context.Context,
	params channels.AuthChannelParameters,
) (*channels.AuthChannelKafka, error) {
	return channels.OpenAuthChannelKafka(
		ctx, params, s,
	)
}

func (s TestServer) OpenAuthRequestOperationKafka(
	ctx context.Context,
	params channels.AuthChannelParameters,
) (*operations.AuthRequestOperationKafka, error) {
	return operations.OpenAuthRequestOperationKafka(
		ctx, params, s,
	)
}
func (s TestServer) OpenAuthResponseOperationKafka(
	ctx context.Context,
	params channels.AuthChannelParameters,
) (*operations.AuthResponseOperationKafka, error) {
	return operations.OpenAuthResponseOperationKafka(
		ctx, params, s,
	)
}
