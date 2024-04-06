+++
title = 'Overview'
weight = 100
+++

# Overview

`go-asyncapi` is the codegen tool that generates the boilerplate Go code based on [AsyncAPI](https://www.asyncapi.com/)
documents. Especially, it is useful for architectures that use the schema-first development approach.

{{< hint info >}}
**[AsyncAPI](https://www.asyncapi.com/)** is a specification for defining APIs for event-driven architectures. The
AsyncAPI document describes the messages, channels, servers, and other entities that the systems in event-driven
architecture use to communicate with each other.
{{< /hint >}}

This tool takes the **servers**, **channels**, **messages**, **models** and other objects from AsyncAPI document and
turns them into Go code. This code is not only just a bunch of stubs, it also contains the schemas and abstract
logic for sending and receiving the messages through the channels and servers depending on protocols you use. The
generated code is modular, so any object can be used separately or be reused from other location.

Also, `go-asyncapi` provides the predefined minimal **implementations** for all supported protocols (at least one
for each protocol), which are attached to the generated code by default.

*Batteries included, but removable* :)

Supported AsyncAPI entities and protocols are listed in the [Features]({{< relref "/docs/features" >}}) page.

## Installation

```bash
go install github.com/bdragon300/go-asyncapi/cmd/go-asyncapi@latest
```

Simple example to begin from is described in the [Quickstart]({{< relref "/docs/quickstart" >}}) page.

## Project versioning

`go-asyncapi` uses [Semantic Versioning](https://semver.org/) for versioning. For example, `1.4.0`.

Releasing a *patch version* contains only bug fixes and minor improvements. You won't need to regenerate the code after
upgrading the tool. E.g. **1.4.0 &rarr; 1.4.1**.

Releasing a *minor version* means that the generated code may be affected, but without breaking changes. You may need to
regenerate the code. E.g. **1.4.0 &rarr; 1.5.0**.

*Major version* release typically introduces the breaking changes. You may need to regenerate the code and to fix your 
projects that uses it. E.g. **1.4.0 &rarr; 2.0.0**.

## FAQ

**Why do I need another third-party codegen tool? We have already the [official generator](https://github.com/asyncapi/generator)**

The official generator is quite specific for most cases. At the moment, it produces the code only for Watermill 
framework, but not everyone uses the Watermill in their projects. Also, it supports only the AMQP protocol.

`go-asyncapi` produces framework-agnostic code with the standard Go library as single dependency. It supports many
[protocols]({{< relref "/docs/features#protocols" >}}) and many specific AsyncAPI entities, such as
bindings, correlation ids, server variables, etc.

`go-asyncapi` contains the simple clients for all supported protocols based on popular libraries. They are modular, so
you can use them directly or as a base for your own implementation. Or don't use them at all, if you don't need them.

And finally, `go-asyncapi` is written in Go, so you don't need node.js or Docker or similar tools to run the generator.

**How can I customize the generated code templates?**

`go-asyncapi` has many ways to customize the generated code, see the command line flags and `x-` fields description.

However, unlike the more common approach for other codegen tools that use templates, the `go-asyncapi` uses the
[jennifer](https://github.com/dave/jennifer) library. This approach is less customizable by user, but more
flexible. It's easier to make support of AsyncAPI specification more complete this way and deal with complex documents 
with plenty of interlinked objects.

But still, user templates in some reduced form is the planned feature.

## Alternatives

* https://github.com/asyncapi/generator (official generator)
* https://github.com/lerenn/asyncapi-codegen
* https://github.com/c0olix/asyncApiCodeGen
