---
title: "Channel"
weight: 1
# bookFlatSection: false
# bookToc: true
# bookHidden: false
# bookCollapseSection: false
# bookComments: false
# bookSearchExclude: false
---

# Channel

## Overview

Channel is an AsyncAPI entity that represents a communication channel using one or more 
[Servers]({{< relref "/docs/code-structure/server" >}}). Channel is protocol-agnostic, the concrete protocol is 
set by the Server object that this channel is bound to. However, the channel may contain protocol-specific properties,
see [Channel bindings](#channel-bindings-operation-bindings).

Typically, a channel may be one-directional or bidirectional. It also may have no direction, which means it acts mostly 
like a stub. This depends on `publish` and `subscribe` sections in document definition.

The generated channel code contains some common methods and fields. Depending on the channel direction, the channel 
code contains proper publisher/subscriber fields and methods to send/receive 
[Message envelopes]({{< relref "/docs/code-structure/message" >}}).

The channel code also contains an interface for servers bound to this channel. All servers, that are bound to this 
channel in the AsyncAPI document are complies to this interface (we automatically append the needed methods to the
server code during generation).

And finally, the channel code contains a convenience method to open the channel for given servers.

By default, the channel code is generated in the `channels` package.

{{< details "Minimal example" >}}
{{< tabs "1" >}}
{{< tab "Definition" >}}
```yaml
channels:
  myChannel:
    description: My channel
    publish:
      message:
        payload:
          type: object
          properties:
            value:
              type: string

servers:
  myServer:
    url: 'kafka://localhost:9092'
    protocol: kafka
```
{{< /tab >}}

{{< tab "Produced channel code" >}}
```go
package channels

import (
	"context"
	"errors"
	run "github.com/bdragon300/go-asyncapi/run"
	kafka "github.com/bdragon300/go-asyncapi/run/kafka"
)

func MyChannelName() run.ParamString {
	return run.ParamString{Expr: "myChannel"}
}

type myChannelKafkaServer interface {
	OpenMyChannelKafka(ctx context.Context) (*MyChannelKafka, error)
	Producer() kafka.Producer
}

func OpenMyChannelKafka(ctx context.Context, servers ...myChannelKafkaServer) (*MyChannelKafka, error) {
	if len(servers) == 0 {
		return nil, run.ErrEmptyServers
	}
	name := MyChannelName()
	var prod []kafka.Producer
	for _, srv := range servers {
		prod = append(prod, srv.Producer())
	}
	pubs, err := run.GatherPublishers[kafka.EnvelopeWriter, kafka.Publisher, kafka.ChannelBindings](ctx, name, nil, prod)

	if err != nil {
		return nil, err
	}
	pub := run.PublisherFanOut[kafka.EnvelopeWriter, kafka.Publisher]{Publishers: pubs}
	ch := NewMyChannelKafka(pub)
	return ch, nil
}
func NewMyChannelKafka(publisher kafka.Publisher) *MyChannelKafka {
	res := MyChannelKafka{
		name:      MyChannelName(),
		publisher: publisher,
	}
	res.topic = res.name.String()
	return &res
}

// MyChannelKafka -- my channel
type MyChannelKafka struct {
	name      run.ParamString
	publisher kafka.Publisher
	topic     string
}

func (m MyChannelKafka) Name() run.ParamString {
	return m.name
}
func (m MyChannelKafka) Close() (err error) {
	if m.publisher != nil {
		err = errors.Join(err, m.publisher.Close())
	}
	return
}
func (m MyChannelKafka) Topic() string {
	return m.topic
}
func (m MyChannelKafka) Publisher() kafka.Publisher {
	return m.publisher
}
func (m MyChannelKafka) Publish(ctx context.Context, envelopes ...kafka.EnvelopeWriter) error {
	return m.publisher.Send(ctx, envelopes...)
}
func (m MyChannelKafka) MakeEnvelope(envelope kafka.EnvelopeWriter, message *MyChannelMessageOut) error {
	envelope.ResetPayload()

	if err := message.MarshalKafkaEnvelope(envelope); err != nil {
		return err
	}
	envelope.SetTopic(m.topic)
	return nil
}

```
{{< /tab >}}

{{< tab "Produced server code" >}}

Pay attention to `OpenMyChannelKafka` method, generated to comply the `myChannelKafkaServer` interface in channel code.

```go
package servers

import (
	"context"
	channels "myproject/channels"
	run "github.com/bdragon300/go-asyncapi/run"
	kafka "github.com/bdragon300/go-asyncapi/run/kafka"
)

func MyServerURL() run.ParamString {
	return run.ParamString{Expr: "kafka://localhost:9092"}
}

type MyServerBindings struct{}

func (m MyServerBindings) Kafka() kafka.ServerBindings {
	b := kafka.ServerBindings{SchemaRegistryURL: "http://localhost:8081"}
	return b
}
func NewMyServer(producer kafka.Producer, consumer kafka.Consumer) *MyServer {
	return &MyServer{
		consumer: consumer,
		producer: producer,
	}
}

type MyServer struct {
	producer kafka.Producer
	consumer kafka.Consumer
}

func (m MyServer) Name() string {
	return "myServer"
}
func (m MyServer) OpenMyChannelKafka(ctx context.Context) (*channels.MyChannelKafka, error) {
	return channels.OpenMyChannelKafka(ctx, m)
}
func (m MyServer) Producer() kafka.Producer {
	return m.producer
}
func (m MyServer) Consumer() kafka.Consumer {
	return m.consumer
}
```
{{< /tab >}}
{{< /tabs >}}
{{< /details >}}

## Document scope

A channel can be defined in two places in the AsyncAPI document:

- `channels` section
- `components.channels` section

Channels in `channels` produce the code. Channels in `components.channels` are just reusable objects, that produce the
code only being referred somewhere in the `channels` section. Therefore, if such channel is not referred anywhere,
it will not be generated at all.

So, in example below, only the `spam` and `eggs` are considered, `fooChannel` will be ignored:

{{< details "Example" >}}
```yaml
channels:
  spam:  # <-- will be generated
    publish:
      message:
        payload:
        type: string
  eggs:  # <-- will be generated
    $ref: '#/components/channels/eggsChannel'
    
components:
  channels:
    eggsChannel:  # <-- will be generated as `eggs` (it's referred by the `eggs` channel)
      subscribe:
        message:
          payload:
            type: string
    fooChannel:  # <-- will NOT be generated (does not appear in the `channels` section)
      publish:
        message:
          payload:
            type: integer
```
{{< /details >}}

In a similar way, only the servers from the `servers` section are considered. See the
[Servers]({{< relref "/docs/code-structure/server" >}}) for more details.

## Operation

### x-ignore

If this extra field it set to **true**, the operation will not be generated. All references
to objects in this operation in the generated code (if any) are replaced by Go `any` type.

{{< details "Example" >}}
```yaml
servers:
  myServer:
    url: 'kafka://localhost:9092'
    protocol: kafka

channels:
  myChannel:
    publish:
      x-ignore: true
      message:
        payload:
          type: string
```
{{< /details >}}

## Channel parameters

Channel parameters are variables, that are substituted in the channel name during channel opening. This is useful when
the channel name is not known at the time of the code generation. For example, when channel name sets the topic or 
queue to use, and you want to determine it at runtime.

The channel parameters are defined in the `parameters` section and are substituted to the appropriate placeholders, 
enclosed in curly braces.

{{< details "Channel parameters example" >}}
{{< tabs "2" >}}
{{< tab "Definition" >}}
```yaml
channels:
  mychannel_{variant}:
    description: My channel
    parameters:
      variant:
        schema:
          type: string
    publish:
      message:
        payload:
          type: object
          properties:
            value:
              type: string
```
{{< /tab >}}

{{< tab "Produced code" >}}

**variant.go**:

```go
package channels

import "fmt"

type Variant struct {
	Value string
}

func (v Variant) Name() string {
	return "variant"
}
func (v Variant) String() string {
	return fmt.Sprint(v.Value)
}
```

**mychannel_variant.go**:

```go
package channels

//...

type MychannelVariantParameters struct {
	Variant Variant
}

func MychannelVariantName(params MychannelVariantParameters) run.ParamString {
	paramMap := map[string]string{params.Variant.Name(): params.Variant.String()}
	return run.ParamString{
		Expr:       "mychannel_{variant}",
		Parameters: paramMap,
	}
}

type mychannelVariantKafkaServer interface {
	OpenMychannelVariantKafka(ctx context.Context, params MychannelVariantParameters) (*MychannelVariantKafka, error)
	Producer() kafka.Producer
}

func OpenMychannelVariantKafka(ctx context.Context, params MychannelVariantParameters, servers ...mychannelVariantKafkaServer) (*MychannelVariantKafka, error) {
//...
}
func NewMychannelVariantKafka(params MychannelVariantParameters, publisher kafka.Publisher) *MychannelVariantKafka {
	res := MychannelVariantKafka{
		name:      MychannelVariantName(params),
		publisher: publisher,
	}
	res.topic = res.name.String()
	return &res
}

//...
```

{{< /tab >}}

{{< tab "Usage" >}}

```go
channelParams := MychannelVariantParameters{
	Variant: Variant{Value: "foo"}
}
channelName := MychannelVariantName(channelParams)
fmt.Println(channelName) 
// Output: mychannel_foo
channel, err := OpenMychannelVariantKafka(context.Background(), channelParams, myServer)
if err != nil {
    log.Fatalf("Failed to open channel: %v", err)
}
defer channel.Close()
```

{{< /tab >}}
{{< /tabs >}}
{{< /details >}}

### x-go-name

Explicitly set the name of the parameter in generated code. By default, the Go name is taken from a parameter name.

{{< details "Example" >}}
{{< tabs "3" >}}
{{< tab "Definition" >}}
```yaml
channels:
  mychannel_{variant}:
    description: My channel
    parameters:
      variant:
        schema:
          type: string
        x-go-name: VariantName
    publish:
      message:
        payload:
          type: object
          properties:
            value:
              type: string
```
{{< /tab >}}
{{< /tabs >}}
{{< /details >}}

## Channel bindings, operation bindings

Channel bindings are protocol-specific properties, that are used to define the channel behavior. They
are defined in the `bindings` section of the channel definition.

{{< details "Channel bindings example" >}}
{{< tabs "3" >}}
{{< tab "Definition" >}}
```yaml
channels:
  mychannel:
    description: My channel
    publish:
      message:
        payload:
          type: object
          properties:
            value:
              type: string
      bindings:
        kafka:
          clientId:  # This should contain jsonschema definition according to AsyncAPI spec 
            type: string
            default: "my-client"
    bindings:
      kafka:
        partitions: 64
```
{{< /tab >}}

{{< tab "Produced code" >}}
```go
package channels

//...

type MychannelBindings struct{}

func (m MychannelBindings) Kafka() kafka.ChannelBindings {
	b := kafka.ChannelBindings{Partitions: 64}
	clientID := "{\"default\":\"my-client\",\"type\":\"string\"}"
	_ = json.Unmarshal([]byte(clientID), &b.PublisherBindings.ClientID)
	return b
}

//...
```
{{< /tab >}}

{{< tab "Usage" >}}

The "Open channel" function automatically uses the bindings, if any:

```go
channel, err := OpenMychannelKafka(context.Background(), myServer)
if err != nil {
    log.Fatalf("Failed to open channel: %v", err)
}
defer channel.Close()
```

At a lower level the channel bindings are used to make a Publisher/Subscriber object:

```go
publisher, err := producer.NewPublisher(ctx, "mychannel", MychannelBindings.Kafka())
if err != nil {
    log.Fatalf("Failed to create Kafka publisher: %v", err)
}
defer publisher.Close()
```

{{< /tab >}}
{{< /tabs >}}
{{< /details >}}

## x-go-name

This extra field is used to explicitly set the name of the channel in generated code. By default, the Go name is 
generated from the AsyncAPI channel name.

{{< details "Example" >}}
{{< tabs "4" >}}
{{< tab "Definition" >}}
```yaml
servers:
  myServer:
    url: 'kafka://localhost:9092'
    protocol: kafka

channels:
  myChannel:
    x-go-name: FooBar
    description: My channel
    publish:
      message:
        payload:
          type: string
```
{{< /tab >}}

{{< tab "Produced code" >}}
```go

//...

// FooBarKafka -- my channel
type FooBarKafka struct {
    name      run.ParamString
    publisher kafka.Publisher
    topic     string
}

//...
```
{{< /tab >}}
{{< /tabs >}}
{{< /details >}}

## x-ignore

If this extra field it set to **true**, the channel will not be generated. All references 
to this channel in the generated code (if any) are replaced by Go `any` type.

{{< details "Example" >}}
```yaml
servers:
  myServer:
    url: 'kafka://localhost:9092'
    protocol: kafka

channels:
  myChannel:
    x-ignore: true
    publish:
      message:
        payload:
          type: string
```
{{< /details >}}