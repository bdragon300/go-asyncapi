+++
title = 'Introduction'
weight = 10
+++

# Introduction

## Overview

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

Supported AsyncAPI entities and protocols are listed in the [Features]({{< relref "/docs/overview/features" >}}) page.

Now only the version 2.6.0 of AsyncAPI specification is supported.

## Installation

```bash
go install github.com/bdragon300/go-asyncapi/cmd/go-asyncapi@latest
```

#TODO: code example
#TODO: versioning

## FAQ

**What the difference between `go-asyncapi` and official [generator](https://github.com/asyncapi/generator)?**

At the moment, the official generator is quite special, it produces the code only for Watermill framework, but not every
project uses the Watermill. Also, it supports only the AMQP protocol.

`go-asyncapi` produces framework-agnostic code with the standard Go library as dependency (except for
pluggable protocol-specific implementations, which is optional). It supports many 
[protocols]({{< relref "/docs/overview/features#protocols" >}}) and many AsyncAPI entities, such as 
bindings, correlation ids, server variables, etc.

And finally, `go-asyncapi` is written in Go, so you don't need node.js or Docker or similar tools to run the generator.

**How can I customize the template that is used to generate the code?**

Unlike the more common approach for codegen tools to use the templates, the `go-asyncapi` uses the 
[jennifer](https://github.com/dave/jennifer) library. This approach is less customizable, but more
flexible, and it's easier to make AsyncAPI specification more complete and deal with complex documents with plenty 
of references and objects.

`go-asyncapi` already has the cli flags to customize the result. But still, the user templates in some reduced form 
is a planned feature.

#TODO: add smth more

## Alternatives

#TODO