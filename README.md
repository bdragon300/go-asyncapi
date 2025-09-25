# go-asyncapi

![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/bdragon300/go-asyncapi)
![GitHub Workflow Status (with branch)](https://img.shields.io/github/actions/workflow/status/bdragon300/go-asyncapi/commit.yml?branch=master)

`go-asyncapi` is a Go implementation of the [AsyncAPI](https://www.asyncapi.com/) specification.

> **[AsyncAPI](https://www.asyncapi.com/)** is a specification for defining APIs for event-driven architectures. The
> AsyncAPI document describes the messages, channels, servers, and other entities that the systems in event-driven
> architecture use to communicate with each other.

[Documentation](https://bdragon300.github.io/go-asyncapi/)

## Core features

|                                                                                                                           | Feature                                                                                                                             | Command               |
|---------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------------------------|-----------------------|
| <img src="https://bdragon300.github.io/go-asyncapi/images/go-logo.svg" style="height: 3em; vertical-align: middle">       | [Generating the Go boilerplate code](https://bdragon300.github.io/go-asyncapi/code-generation/overview)                             | `go-asyncapi code`    |
| <img src="https://bdragon300.github.io/go-asyncapi/images/terminal-icon.svg" style="height: 3em; vertical-align: middle"> | [Building the zero-code CLI client executable](https://bdragon300.github.io/go-asyncapi/client-application-generation)              | `go-asyncapi client`  |
| <img src="https://bdragon300.github.io/go-asyncapi/images/infra.svg" style="height: 3em; vertical-align: middle">         | [Generating the Infrastructure-As-Code (IaC) definitions](https://bdragon300.github.io/go-asyncapi/infrastructure-files-generation) | `go-asyncapi infra`   |
| <img src="https://bdragon300.github.io/go-asyncapi/images/diagram.svg" style="height: 3em; vertical-align: middle">       | [Generating the diagrams](https://bdragon300.github.io/go-asyncapi/visualization/)                                                  | `go-asyncapi diagram` |

Also, `go-asyncapi` contains the built-in **protocol implementations** based on popular libraries, supports the **result customization** using Go templates and more.

See the [Features](https://bdragon300.github.io/go-asyncapi/features) page for more details.

## Supported protocols

|                                                                                                                                         | Protocol       | Library                                                                            |
|-----------------------------------------------------------------------------------------------------------------------------------------|----------------|------------------------------------------------------------------------------------|
| <img alt="AMQP" src="https://bdragon300.github.io/go-asyncapi/images/amqp.svg" style="height: 1.5em; vertical-align: middle">           | AMQP           | [github.com/rabbitmq/amqp091-go](https://github.com/rabbitmq/amqp091-go)           |
| <img alt="HTTP" src="https://bdragon300.github.io/go-asyncapi/images/http.svg" style="height: 1.5em; vertical-align: middle">           | HTTP           | [net/http](https://pkg.go.dev/net/http)                                            |
| <img alt="IP RAW Sockets" src="https://bdragon300.github.io/go-asyncapi/images/ip.png" style="height: 1.5em; vertical-align: middle">   | IP RAW Sockets | [net](https://pkg.go.dev/net)                                                      |
| <img alt="Apache Kafka" src="https://bdragon300.github.io/go-asyncapi/images/kafka.svg" style="height: 1.5em; vertical-align: middle">  | Apache Kafka   | [github.com/twmb/franz-go](https://github.com/twmb/franz-go)                       |
| <img alt="NATS" src="https://bdragon300.github.io/go-asyncapi/images/nats.svg" style="height: 1.5em; vertical-align: middle">           | NATS           | [github.com/nats-io/nats.go](https://github.com/nats-io/nats.go)                   |
| <img alt="MQTT" src="https://bdragon300.github.io/go-asyncapi/images/mqtt.svg" style="height: 1.5em; vertical-align: middle">           | MQTT           | [github.com/eclipse/paho.mqtt.golang](https://github.com/eclipse/paho.mqtt.golang) |
| <img alt="Redis" src="https://bdragon300.github.io/go-asyncapi/images/redis.svg" style="height: 1.5em; vertical-align: middle">         | Redis          | [github.com/redis/go-redis](https://github.com/redis/go-redis)                     |
| <img alt="TCP" src="https://bdragon300.github.io/go-asyncapi/images/tcpudp.svg" style="height: 1.5em; vertical-align: middle">          | TCP            | [net](https://pkg.go.dev/net)                                                      |
| <img alt="UDP" src="https://bdragon300.github.io/go-asyncapi/images/tcpudp.svg" style="height: 1.5em; vertical-align: middle">          | UDP            | [net](https://pkg.go.dev/net)                                                      |
| <img alt="Websocket" src="https://bdragon300.github.io/go-asyncapi/images/websocket.svg" style="height: 1.5em; vertical-align: middle"> | Websocket      | [github.com/gobwas/ws](https://github.com/gobwas/ws)                               |

## Usage

The following couple of high-level examples show how to use the generated code for sending and receiving messages.

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
if err := myChannel.OperationFooBar().PublishMyMessage(ctx, msg); err != nil {
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
err := myChannel.OperationFooBar().SubscribeMyMessage(ctx, func(msg messages.MyMessage) {
	log.Printf("received message: %+v", msg)
})
if err != nil {
	log.Fatalf("subscribe: %v", err)
}
```

The low-level functions are also available, which gives more control over the process.

Also, here are some demo applications:

* [Site authorization](https://github.com/bdragon300/go-asyncapi/blob/master/examples/site-authorization)
* [HTTP echo server](https://github.com/bdragon300/go-asyncapi/blob/master/examples/http-server)

## Installation

```bash
go install github.com/bdragon300/go-asyncapi/cmd/go-asyncapi@latest
```

## Project status

The project is in active development. The main features are implemented, but there are still some missing features and
bugs.

This version supports the latest AsyncAPI v3. It doesn't support the previous v2, because v3 introduced some breaking 
changes, so making the tool that supports both v2 and v3 requires a lot of effort and time with no significant benefits.

## Project versioning

`go-asyncapi` uses [Semantic Versioning](https://semver.org/) for versioning. For example, `1.4.0`.

Releasing a *patch version* contains only bug fixes and minor improvements. You won't need to regenerate the code after
upgrading the tool. E.g. **1.4.0 &rarr; 1.4.1**.

Releasing a *minor version* means that the generated code may be affected, but without breaking changes. You may need to
regenerate the code. E.g. **1.4.0 &rarr; 1.5.0**.

*Major version* release typically introduces the breaking changes. You may need to regenerate the code, to fix your
projects that uses it or to change the tool command line. E.g. **1.4.0 &rarr; 2.0.0**.

*Note, that the project major version 0 (0.x.x) is considered unstable*

## Description

### Goals

The goal of the project is, on the one hand, to fully implement the AsyncAPI specification and, on the other hand, to help
the developers, DevOps, QA, and other engineers with making the software in event-driven architectures based on the
AsyncAPI specification.

Another goal is to provide a way to readily use and test the technologies described in the AsyncAPI document —
everything works out of the box, but each component is optional and can be replaced or omitted.

In other words, *batteries included, but removable*.

### Features overview

`go-asyncapi` supports most of the AsyncAPI features, such as messages, channels, servers, bindings, correlation ids, etc.

The generated **Go boilerplate code** has minimal dependencies on external libraries and contains the basic logic sufficient to
send and receive messages. You also can plug in the protocol implementations built-in in `go-asyncapi`, they are based on
popular libraries for that protocol. Also, it is possible to import the third-party code in the code being generated.

It is possible to build the **no-code client application** solely based on the AsyncAPI document, which is useful for
testing purposes or for quick prototyping.

The `go-asyncapi` is able to generate the **infrastructure setup files**, such as Docker Compose files, which are useful
for setting up the development environment quickly or as the starting point for the infrastructure-as-code deploy configurations.

## FAQ

### Why do I need another codegen tool? We already have the [official generator](https://github.com/asyncapi/generator)

Well, `go-asyncapi` provides more features, and it's written in Go.

The official generator is quite specific for many use cases. At the moment, it produces the Go code bound with the
[Watermill](https://watermill.io/) framework, but not everyone uses the Watermill in
their projects. Moreover, a project may have a fixed set of dependencies, for example,
due to the security policies in the company.

Also, the official generator supports only the AMQP protocol.

Instead, `go-asyncapi`:

* produces framework-agnostic code.
* supports more
  [protocols](https://bdragon300.github.io/go-asyncapi/features#protocols) and more specific AsyncAPI entities, such as
  bindings, correlation ids, server variables, etc.
* has built-in clients for all supported protocols based on popular libraries, that may be used in the generated code.
* is written in Go, so no need to have node.js or Docker or similar tools to run the generator.
* can produce IaC files and build the no-code client application.

*Another reason is that I don't know JavaScript well. And I'm not sure that if we want to support all AsyncAPI features,
the existing templates would not be rewritten from the ground.*

### How to contribute?

Just open an issue or a pull request in the [GitHub repository](https://github.com/bdragon300/go-asyncapi)

## Alternatives

* https://github.com/asyncapi/generator (official generator)
* https://github.com/lerenn/asyncapi-codegen
* https://github.com/c0olix/asyncApiCodeGen
