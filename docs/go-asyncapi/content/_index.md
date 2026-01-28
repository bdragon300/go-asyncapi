+++
title = "go-asyncapi documentation"
weight = 1
draft = false
description = 'Go-asyncapi is a tool for working with AsyncAPI documents'
+++

![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/bdragon300/go-asyncapi)
![GitHub Workflow Status (with branch)](https://img.shields.io/github/actions/workflow/status/bdragon300/go-asyncapi/commit.yml?branch=master)

**go-asyncapi** is a tool for working with [AsyncAPI](https://www.asyncapi.com/) documents.

> **[AsyncAPI](https://www.asyncapi.com/)** is a specification for defining APIs for event-driven architectures. The
> AsyncAPI document describes the messages, channels, servers, and other entities that the systems in event-driven
> architecture use to communicate with each other.

[Documentation](https://bdragon300.github.io/go-asyncapi/)

## Core features

|                                                                                 | Feature                                                                                | Command               |
|---------------------------------------------------------------------------------|----------------------------------------------------------------------------------------|-----------------------|
| {{< figure src="images/go-logo.svg" alt="Go code" class="feature-icon">}}       | [Generating the Go boilerplate code for any protocol]({{< relref "/commands/code" >}}) | `go-asyncapi code`    |
| {{< figure src="images/terminal-icon.svg" alt="CLI app" class="feature-icon">}} | [Building the zero-code CLI client executable]({{< relref "/commands/client" >}})      | `go-asyncapi client`  |
| {{< figure src="images/infra.svg" alt="IaC definitions" class="feature-icon">}} | [Generating the server definitions]({{< relref "/commands/infra" >}})                  | `go-asyncapi infra`   |
| {{< figure src="images/diagram.svg" alt="Diagram" class="feature-icon">}}       | [Drawing the SVG diagrams]({{< relref "/commands/diagram" >}})                         | `go-asyncapi diagram` |
| {{< figure src="images/ui-icon.svg" alt="UI" class="feature-icon">}}            | [Serving or generating the web docs UI]({{< relref "/commands/ui" >}})                 | `go-asyncapi ui`      |

See the [Features](https://bdragon300.github.io/go-asyncapi/features) page for more details.

## Supported protocols

`go-asyncapi` is able to generate the abstract code for any protocol, but for these protocols it also adds the supporting code:

|                                                                             | Protocol       | Implementation library                                                             |
|-----------------------------------------------------------------------------|----------------|------------------------------------------------------------------------------------|
| {{< figure src="images/amqp.svg" alt="AMQP" class="brand-icon">}}           | AMQP           | [github.com/rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go)           |
| {{< figure src="images/http.svg" alt="HTTP" class="brand-icon">}}           | HTTP           | [net/http](https://pkg.go.dev/net/http)                                            |
| {{< figure src="images/ip.png" alt="IP RAW Sockets" class="brand-icon">}}   | IP RAW Sockets | [net](https://pkg.go.dev/net)                                                      |
| {{< figure src="images/kafka.svg" alt="Apache Kafka" class="brand-icon">}}  | Apache Kafka   | [github.com/twmb/franz-go](https://github.com/twmb/franz-go)                       |
| {{< figure src="images/mqtt.svg" alt="MQTT v3" class="brand-icon">}}        | MQTT v3        | [github.com/eclipse/paho.mqtt.golang](https://github.com/eclipse/paho.mqtt.golang) |
| {{< figure src="images/mqtt.svg" alt="MQTT v5" class="brand-icon">}}        | MQTT v5        | [github.com/eclipse/paho.golang](https://github.com/eclipse-paho/paho.golang)      |
| {{< figure src="images/nats.svg" alt="NATS" class="brand-icon">}}           | NATS           | [github.com/nats-io/nats.go](https://github.com/nats-io/nats.go)                   |
| {{< figure src="images/redis.svg" alt="Redis" class="brand-icon">}}         | Redis          | [github.com/redis/go-redis](https://github.com/redis/go-redis)                     |
| {{< figure src="images/tcpudp.svg" alt="TCP" class="brand-icon">}}          | TCP            | [net](https://pkg.go.dev/net)                                                      |
| {{< figure src="images/tcpudp.svg" alt="UDP" class="brand-icon">}}          | UDP            | [net](https://pkg.go.dev/net)                                                      |
| {{< figure src="images/websocket.svg" alt="WebSocket" class="brand-icon">}} | WebSocket      | [github.com/gobwas/ws](https://github.com/gobwas/ws)                               |

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

### Goals

The goal of the project is to help the developers, DevOps, QA, and other engineers with making the software in 
event-driven architectures based on the AsyncAPI.

Another goal is to provide a way to prototype and test everything described in AsyncAPI document without coding.

The third goal is full support of the AsyncAPI specification and to make it available to use with modern protocols.

### Features overview

`go-asyncapi` supports most of the AsyncAPI features, such as messages, channels, servers, bindings, correlation ids, etc.

The generated **Go boilerplate code** has minimal dependencies on external libraries and contains the basic logic sufficient to
send and receive messages. You also can plug in the protocol implementations built-in in `go-asyncapi`, they are based on
popular libraries for that protocol. Also, it is possible to import the third-party code in the code being generated.

It is possible to build the **no-code client application** solely based on the AsyncAPI document, which is useful for
testing purposes or for quick prototyping.

The `go-asyncapi` is able to generate the **infrastructure setup files**, such as Docker Compose files, which are useful
for setting up the development environment quickly or as the starting point for the infrastructure-as-code deploy configurations.

Another major feature is drawing the **diagrams** showing the relationships between things described in one or more
AsyncAPI documents. This helps to visualize the architecture.

**Web UI** feature generates the AsyncAPI web docs. Using the [AsyncAPI React Component](https://github.com/asyncapi/asyncapi-react),
`go-asyncapi` is able to generate a static HTML docs page or even serve it just in one command.

## FAQ

### Why do I need another codegen tool? We already have the [official generator](https://github.com/asyncapi/generator)

Well, `go-asyncapi` provides more features, and it's written in Go.

The official generator is quite specific for many use cases. At the moment, it produces the Go code bound with the
[Watermill](https://watermill.io/) framework, but not everyone uses the Watermill in
their projects. Furthermore, a project may have a fixed set of dependencies, for example,
due to the security policies in the company.

Also, the official generator supports only the AMQP protocol.

Instead, `go-asyncapi`:

* produces the framework-agnostic code for *any protocol* and additional supporting code for built-in
  [protocols](https://bdragon300.github.io/go-asyncapi/docs/features#protocols).
* besides the codegen feature, it can 
  [build client application](https://bdragon300.github.io/go-asyncapi/docs/commands/client),
  [draw diagrams](https://bdragon300.github.io/go-asyncapi/docs/commands/diagram), 
  [generate server definitions](https://bdragon300.github.io/go-asyncapi/docs/command/infra),
  [produce web docs](https://bdragon300.github.io/go-asyncapi/docs/command/ui).
* supports some specific AsyncAPI entities, such as protocol bindings, correlation ids, server variables, etc.
* has built-in clients for supported protocols, that are based on popular libraries.
* written in Go, so no need to have node.js or Docker or similar tools to run the generator.

*Another reason is that I don't know JavaScript well. And I'm not sure that if we want to support all AsyncAPI features,
the existing templates would not be rewritten from the ground.*

### How to contribute?

Just open an issue or a pull request. Branches `master` is the current release, `dev` is for development (next release).

## Alternatives

* https://github.com/asyncapi/generator (official generator)
* https://github.com/lerenn/asyncapi-codegen
* https://github.com/c0olix/asyncApiCodeGen
