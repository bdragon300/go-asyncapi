# go-asyncapi

![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/bdragon300/go-asyncapi)
![GitHub Workflow Status (with branch)](https://img.shields.io/github/actions/workflow/status/bdragon300/go-asyncapi/commit.yml?branch=master)

`go-asyncapi` is a tool for working with [AsyncAPI](https://www.asyncapi.com/) documents.

> **[AsyncAPI](https://www.asyncapi.com/)** is a specification for defining APIs for event-driven architectures. The
> AsyncAPI document describes the messages, channels, servers, and other entities that the systems in event-driven
> architecture use to communicate with each other.

[Documentation](https://bdragon300.github.io/go-asyncapi/)

## Core features

|                                                                                                                           | Feature                                                                                                               | Command                        |
|---------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------|--------------------------------|
| <img src="https://bdragon300.github.io/go-asyncapi/images/go-logo.svg" style="height: 3em; vertical-align: middle">       | [Generating the Go boilerplate code for any protocol](https://bdragon300.github.io/go-asyncapi/commands/code)         | `go-asyncapi code`             |
| <img src="https://bdragon300.github.io/go-asyncapi/images/terminal-icon.svg" style="height: 3em; vertical-align: middle"> | [Building the zero-code CLI client executable](https://bdragon300.github.io/go-asyncapi/commands/client)              | `go-asyncapi client`           |
| <img src="https://bdragon300.github.io/go-asyncapi/images/infra.svg" style="height: 3em; vertical-align: middle">         | [Generating the server definitions](https://bdragon300.github.io/go-asyncapi/commands/infra)                          | `go-asyncapi infra`            |
| <img src="https://bdragon300.github.io/go-asyncapi/images/diagram.svg" style="height: 3em; vertical-align: middle">       | [Drawing the SVG diagrams](https://bdragon300.github.io/go-asyncapi/commands/diagram)                                 | `go-asyncapi diagram`          |
| <img src="https://bdragon300.github.io/go-asyncapi/images/ui-icon.svg" style="height: 3em; vertical-align: middle">       | [Serving or generating the web docs UI](https://bdragon300.github.io/go-asyncapi/commands/ui)                         | `go-asyncapi ui`               |
| <img src="https://bdragon300.github.io/go-asyncapi/images/doc-icon.svg" style="height: 3em; vertical-align: middle">      | [Validating the AsyncAPI documents](https://bdragon300.github.io/go-asyncapi/commands/doc/validate)                   | `go-asyncapi doc validate`     |
| <img src="https://bdragon300.github.io/go-asyncapi/images/doc-icon.svg" style="height: 3em; vertical-align: middle">      | [Inspecting the document structure](https://bdragon300.github.io/go-asyncapi/commands/doc/inspect)                    | `go-asyncapi doc inspect`      |
| <img src="https://bdragon300.github.io/go-asyncapi/images/doc-icon.svg" style="height: 3em; vertical-align: middle">      | [Copying the nodes](https://bdragon300.github.io/go-asyncapi/commands/doc/cp)                                         | `go-asyncapi doc cp`           |
| <img src="https://bdragon300.github.io/go-asyncapi/images/doc-icon.svg" style="height: 3em; vertical-align: middle">      | [Moving the nodes](https://bdragon300.github.io/go-asyncapi/commands/doc/mv)                                          | `go-asyncapi doc mv`           |
| <img src="https://bdragon300.github.io/go-asyncapi/images/doc-icon.svg" style="height: 3em; vertical-align: middle">      | [Generating the synthetic examples for entities](https://bdragon300.github.io/go-asyncapi/commands/doc/gen-examples)  | `go-asyncapi doc gen-examples` |
| <img src="https://bdragon300.github.io/go-asyncapi/images/doc-icon.svg" style="height: 3em; vertical-align: middle">      | [Flattening the AsyncAPI documents](https://bdragon300.github.io/go-asyncapi/commands/doc/flatten)                    | `go-asyncapi doc flatten`      |
| <img src="https://bdragon300.github.io/go-asyncapi/images/doc-icon.svg" style="height: 3em; vertical-align: middle">      | [Showing the documents linked by $refs](https://bdragon300.github.io/go-asyncapi/commands/doc/tree)                   | `go-asyncapi doc tree`         |

See the [Features](https://bdragon300.github.io/go-asyncapi/features) page for more details.

## Supported protocols

`go-asyncapi` is able to generate the abstract code for any protocol, but for these protocols it also adds the supporting code: 

|                                                                                                                                         | Protocol       | Implementation library                                                             |
|-----------------------------------------------------------------------------------------------------------------------------------------|----------------|------------------------------------------------------------------------------------|
| <img alt="AMQP" src="https://bdragon300.github.io/go-asyncapi/images/amqp.svg" style="height: 1.5em; vertical-align: middle">           | AMQP           | [github.com/rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go)           |
| <img alt="HTTP" src="https://bdragon300.github.io/go-asyncapi/images/http.svg" style="height: 1.5em; vertical-align: middle">           | HTTP           | [net/http](https://pkg.go.dev/net/http)                                            |
| <img alt="IP RAW Sockets" src="https://bdragon300.github.io/go-asyncapi/images/ip.png" style="height: 1.5em; vertical-align: middle">   | IP RAW Sockets | [net](https://pkg.go.dev/net)                                                      |
| <img alt="Apache Kafka" src="https://bdragon300.github.io/go-asyncapi/images/kafka.svg" style="height: 1.5em; vertical-align: middle">  | Apache Kafka   | [github.com/twmb/franz-go](https://github.com/twmb/franz-go)                       |
| <img alt="MQTT v3" src="https://bdragon300.github.io/go-asyncapi/images/mqtt.svg" style="height: 1.5em; vertical-align: middle">        | MQTT v3        | [github.com/eclipse/paho.mqtt.golang](https://github.com/eclipse/paho.mqtt.golang) |
| <img alt="MQTT v5" src="https://bdragon300.github.io/go-asyncapi/images/mqtt.svg" style="height: 1.5em; vertical-align: middle">        | MQTT v5        | [github.com/eclipse/paho.golang](https://github.com/eclipse-paho/paho.golang)      |
| <img alt="NATS" src="https://bdragon300.github.io/go-asyncapi/images/nats.svg" style="height: 1.5em; vertical-align: middle">           | NATS           | [github.com/nats-io/nats.go](https://github.com/nats-io/nats.go)                   |
| <img alt="Redis" src="https://bdragon300.github.io/go-asyncapi/images/redis.svg" style="height: 1.5em; vertical-align: middle">         | Redis          | [github.com/redis/go-redis](https://github.com/redis/go-redis)                     |
| <img alt="TCP" src="https://bdragon300.github.io/go-asyncapi/images/tcpudp.svg" style="height: 1.5em; vertical-align: middle">          | TCP            | [net](https://pkg.go.dev/net)                                                      |
| <img alt="UDP" src="https://bdragon300.github.io/go-asyncapi/images/tcpudp.svg" style="height: 1.5em; vertical-align: middle">          | UDP            | [net](https://pkg.go.dev/net)                                                      |
| <img alt="Websocket" src="https://bdragon300.github.io/go-asyncapi/images/websocket.svg" style="height: 1.5em; vertical-align: middle"> | Websocket      | [github.com/gobwas/ws](https://github.com/gobwas/ws)                               |


## Installation

```bash
go install github.com/bdragon300/go-asyncapi/cmd/go-asyncapi@latest
```

## Usage

Demo applications:

* [Site authorization](https://github.com/bdragon300/go-asyncapi/blob/master/examples/site-authorization)
* [HTTP echo server](https://github.com/bdragon300/go-asyncapi/blob/master/examples/http-server)

These couple of high-level examples show how to use the generated code for sending and receiving messages.
The low-level functions are also available, which gives more control over the process.

*Publishing*:

```go
ctx := context.Background()
// Connect to broker for sending messages
myServer, err := servers.ConnectMyServerProducer(ctx, servers.MyServerURL())
if err != nil {
	log.Fatalf("connect to the myServer: %v", err)
}
defer myServer.Close()

// Open an channel for sending messages
myChannel, err := myServer.OpenMyChannelKafka(ctx)
if err != nil {
	log.Fatalf("open myChannel: %v", err)
}
defer myChannel.Close()

// Craft a message
msg := messages.MyMessage{
	Payload: schemas.MyMessagePayload{
		Field1: "value1", 
		Field2: 42,
	},
	Headers: schemas.MyMessageHeaders{
		Header1: "header1",
	},
}

// Send a message
if err := myChannel.PublishMyMessage(ctx, msg); err != nil {
	log.Fatalf("send message: %v", err)
}
```

*Subscribing*:

```go
ctx := context.Background()
// Connect to broker for receiving messages
myServer, err := servers.ConnectMyServerConsumer(ctx, servers.MyServerURL())
if err != nil {
	log.Fatalf("connect to the myServer: %v", err)
}
defer myServer.Close()

// Open an channel for sending messages
myChannel, err := myServer.OpenMyChannelKafka(ctx)
if err != nil {
    log.Fatalf("open myChannel: %v", err)
}
defer myChannel.Close()

// Subscribe to messages
err := myChannel.SubscribeMyMessage(ctx, func(msg messages.MyMessage) {
	log.Printf("received message: %+v", msg)
})
if err != nil {
	log.Fatalf("subscribe: %v", err)
}
```

## Project status

The project is in active development and is considered unstable. API may change.

## Description

`go-asyncapi` trying to cover most use cases in the lifecycle of an AsyncAPI-based project with a single tool: 
prototyping, developing, setting up the infrastructure, doing QA and keeping the documents themselves in order. 
Two things underlie all of it: full support of the AsyncAPI specification and usability with modern protocols.

Although the tool is useful for any role involved in an event-driven project (DevOps, QA, architects), 
it is primarily targeted at Go backend developers.

### Why another tool? We have the official generator and other tools already.

Well, `go-asyncapi` provides more features, and it's written in Go, so it doesn't require Node.js or Docker to run.

The official generator is quite specific for many use cases. At the moment, it produces the Go code bound with the
[Watermill](https://watermill.io/) framework, but not everyone uses the Watermill in
their projects. Furthermore, a project may have a fixed set of dependencies, for example,
due to the security policies in the company.

Also, the official generator supports only the AMQP protocol.

Instead, `go-asyncapi`:

* produces the framework-agnostic code and have support for built-in
  [protocols](https://bdragon300.github.io/go-asyncapi/features#protocols). Any protocol can be added to generator
  without modifying the generator - only by writing Go templates.
* besides the codegen feature, it can
  [build client application](https://bdragon300.github.io/go-asyncapi/commands/client),
  [draw diagrams](https://bdragon300.github.io/go-asyncapi/commands/diagram),
  [generate server definitions](https://bdragon300.github.io/go-asyncapi/commands/infra),
  [produce web docs](https://bdragon300.github.io/go-asyncapi/commands/ui),
  [manipulate the AsyncAPI documents](https://bdragon300.github.io/go-asyncapi/commands/doc).
* it supports some specific AsyncAPI entities, such as protocol bindings, correlation ids, server variables, etc.
* has built-in clients for supported protocols, that are based on popular libraries.

*Another reason is that I don't know JavaScript well. And I'm not sure that if we want to support all AsyncAPI features,
the existing templates would not be rewritten from the ground.*

## How to contribute?

Just open an issue or a pull request. Branches `master` is the current release, `dev` is for development (next release).

## Alternatives

* https://github.com/asyncapi/generator (official generator)
* https://github.com/asyncapi/EDAVisualiser (official visualizer)
* https://github.com/asyncapi/bundler (official tool for merging AsyncAPI documents)
* https://github.com/asyncapi/asyncapi-react (official UI component for rendering AsyncAPI documentation)
* https://github.com/lerenn/asyncapi-codegen
