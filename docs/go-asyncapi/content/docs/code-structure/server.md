---
title: "Server"
weight: 410
# bookFlatSection: false
# bookToc: true
# bookHidden: false
# bookCollapseSection: false
# bookComments: false
# bookSearchExclude: false

---

# Server

## Overview

Server is one of the main entities in the AsyncAPI document. It represents a server that is used by the channels to
send and receive messages. Server is always assigned to a particular protocol, such as Kafka, AMQP, MQTT, etc. Also, it
always has a URL to connect to.

The generated server code contains some common methods and fields, and also method to open channels that are bound to
this server. By default, the server code is generated in the `servers` package.

To learn how a server is reflected in implementation code, see the 
[Implementations]({{< relref "/docs/code-structure/implementation#server--producerconsumer" >}}) page.

{{< details "Minimal example" >}}
{{< tabs "1" >}}
{{< tab "Definition" >}}
```yaml
channels:
  myChannel:
    description: My channel

servers:
  myServer:
    url: 'kafka://localhost:9092'
    protocol: kafka
```
{{< /tab >}}

{{< tab "Produced code" >}}
```go
package servers

func MyServerURL() run.ParamString {
  return run.ParamString{Expr: "kafka://localhost:9092"}
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

A server can be defined in two sections in the AsyncAPI document:

- `servers`
- `components.servers`

Servers in `servers` produce the code. Servers in `components.servers` are just reusable objects, that produce the
code only being referred somewhere in the `servers` section. Therefore, if such server is not referred anywhere, 
it will not be generated at all.

So, in example below, only the `foo` and `bar` are considered, `spam` will be ignored: 

{{< details "Example" >}}
```yaml
servers:
  foo:  # <-- will be generated
    url: 'kafka://localhost:9092'
    protocol: kafka
  bar:  # <-- will be generated
    $ref: '#/components/servers/barServer'
    
components:
  servers:
    barServer:  # <-- will be generated as `bar` (it's referred by the `bar` server)
      url: 'mqtt://localhost:1883'
      protocol: mqtt
    spamServer:  # <-- will NOT be generated (does not appear in the `servers` section)
      url: 'amqp://localhost:5672'
      protocol: amqp
```
{{< /details >}}

In a similar way, only the channels from the `channels` section are considered for the server code generation. See the
[Channels]({{< relref "/docs/code-structure/channel" >}}) for more details.

## Server variables

Server variables are used for the server URL templating. They are defined in the `variables` section of the server 
object and are substituted to the appropriate placeholders, enclosed in curly braces.

{{< details "Server variables example" >}}
{{< tabs "2" >}}
{{< tab "Definition" >}}
```yaml
servers:
  myServer:
    url: 'kafka://{host}:{port}'
    protocol: kafka
    variables:
      host:
        default: 'localhost'
      port:
        default: '9092'
```
{{< /tab >}}

{{< tab "Produced code" >}}

```go
package servers

func MyServerURL(host string, port string) run.ParamString {
  paramMap := map[string]string{"host": host, "port": port}
  return run.ParamString{Expr: "kafka://{host}:{port}", ParamMap: paramMap}
}

// ...
```
{{< /tab >}}

{{< tab "Usage" >}}

`ParamString` complies the `fmt.Stringer` interface, so you can use it as a string:

```go 
fmt.Println(MyServerURL("localhost", "9092"))
// Output: kafka://localhost:9092
```
{{< /tab >}}
{{< /tabs >}}
{{< /details >}}

## Server bindings

Server bindings set the protocol-specific properties for the server. They are defined in the `bindings` section of the
server object.

{{< details "Server bindings example" >}}
{{< tabs "3" >}}
{{< tab "Definition" >}}
```yaml
servers:
  myServer:
    url: 'kafka://localhost:9092'
    protocol: kafka
    bindings:
      kafka:
        schemaRegistryUrl: 'http://localhost:8081'
```
{{< /tab >}}

{{< tab "Produced code" >}}
```go
package servers

//...

type MyServerBindings struct{}

func (m MyServerBindings) Kafka() kafka.ServerBindings {
  b := kafka.ServerBindings{SchemaRegistryURL: "http://localhost:8081"}
  return b
}

//...
```
{{< /tab >}}

{{< tab "Usage" >}}
Typically, bindings are passed to implementation to set Consumer or Producer creation options:

```go
var consumer = implKafka.NewConsumer(MyServerURL().String(), MyServerBindings().Kafka(), nil) 
```
{{< /tab >}}
{{< /tabs >}}
{{< /details >}}

## x-go-name

This extra field is used to explicitly set the name of the server in generated code. By default, the Go name is
generated from the AsyncAPI server name.

{{< details "Example" >}}
{{< tabs "4" >}}
{{< tab "Definition" >}}
```yaml
channels:
  myChannel:
    description: My channel

servers:
  myServer:
    url: 'kafka://localhost:9092'
    protocol: kafka
    x-go-name: FooBar
```
{{< /tab >}}

{{< tab "Produced code" >}}
```go

//...

type FooBar struct {
    producer kafka.Producer
    consumer kafka.Consumer
}

//...
```
{{< /tab >}}
{{< /tabs >}}
{{< /details >}}

## x-ignore

If this extra field it set to **true**, the server will not be generated. All references
to this server in the generated code (if any) are replaced by Go `any` type.

{{< details "Example" >}}
```yaml
channels:
  myChannel:
    description: My channel

servers:
  myServer:
    url: 'kafka://localhost:9092'
    protocol: kafka
    x-ignore: true
```
{{< /details >}}