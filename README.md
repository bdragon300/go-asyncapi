# go-asyncapi

![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/bdragon300/go-asyncapi)
![GitHub Workflow Status (with branch)](https://img.shields.io/github/actions/workflow/status/bdragon300/go-asyncapi/go-test.yml?branch=master)

`go-asyncapi` is a Go implementation of the [AsyncAPI](https://www.asyncapi.com/) specification.

> **[AsyncAPI](https://www.asyncapi.com/)** is a specification for defining APIs for event-driven architectures. The
> AsyncAPI document describes the messages, channels, servers, and other entities that the systems in event-driven
> architecture use to communicate with each other.

[Documentation](https://bdragon300.github.io/go-asyncapi/)

## Core features

Based on an AsyncAPI document, with `go-asyncapi` you can:

* Generate the boilerplate code for your project
* Build the no-code client application (e.g. for testing purposes)
* Generate the server infrastructure setup files (e.g. docker-compose files)
* Use built-in protocol implementations (AMQP, Kafka, WebSocket, etc.)

For more details, see the [Features](https://bdragon300.github.io/go-asyncapi/docs/features) page.

## Description

### Goals

The goal of the project is, on the one hand, to fully implement the AsyncAPI specification and, on the other hand, to help
the developers, DevOps, QA, and other engineers with developing the software in event-driven architectures based on the 
AsyncAPI specification.

Another goal is to provide a way to plug in the ready-to-use technologies described in the AsyncAPI
document to a project without losing the flexibility. In other words, *batteries are included, but removable*.

### Features overview

`go-asyncapi` supports most of the AsyncAPI features, such as messages, channels, servers, bindings, correlation ids, etc.

The generated Go boilerplate code has no dependencies on external libraries and contains the basic logic sufficient to 
send and receive messages. You also can plug in the protocol implementations built-in in `go-asyncapi`, they are based on 
popular libraries for that protocol. Also, it is possible to substitute the code generated before.

It is possible to build the no-code client application solely based on the AsyncAPI document, which is useful for
testing purposes or for quick prototyping.

The `go-asyncapi` is able to generate the infrastructure setup files, such as Docker Compose files, which are useful
to set up the development environment quickly or as the starting point for the infrastructure-as-code deploy configurations.

The tool also supports both the internal and external `$ref` references. The latter may point to the local files or to 
the remote documents available via HTTP(S) or you can provide a custom resolver.

The behavior can be customized using the command-line flags, a YAML config file, and `x-` fields in the AsyncAPI document.
`go-asyncapi` also accepts the user Go templates to customize the output.

## Installation

```bash
go install github.com/bdragon300/go-asyncapi/cmd/go-asyncapi@latest
```

## Project versioning

`go-asyncapi` uses [Semantic Versioning](https://semver.org/) for versioning. For example, `1.4.0`.

Releasing a *patch version* contains only bug fixes and minor improvements. You won't need to regenerate the code after
upgrading the tool. E.g. **1.4.0 &rarr; 1.4.1**.

Releasing a *minor version* means that the generated code may be affected, but without breaking changes. You may need to
regenerate the code. E.g. **1.4.0 &rarr; 1.5.0**.

*Major version* release typically introduces the breaking changes. You may need to regenerate the code, to fix your
projects that uses it or to change the tool command line. E.g. **1.4.0 &rarr; 2.0.0**.

*Note, that the project major version 0 (0.x.x) is considered unstable*

## FAQ

### Why do I need another codegen tool? We already have the [official generator](https://github.com/asyncapi/generator)

**TLDR**: `go-asyncapi` provides more features, and it's written in Go.

The official generator is quite specific for many use cases. At the moment, it produces the Go code bound with the
[Watermill](https://watermill.io/) framework, but not everyone uses the Watermill in
their projects. Moreover, a project may have a fixed set of dependencies, for example,
due to the security policies in the company.

Also, the official generator supports only the AMQP protocol.

Instead, `go-asyncapi`:

* produces framework-agnostic code.
* supports more
  [protocols](https://bdragon300.github.io/go-asyncapi/docs/features#protocols) and more specific AsyncAPI entities, such as
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
