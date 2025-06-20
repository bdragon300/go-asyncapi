+++
title = 'Features'
bookCollapseSection = true
weight = 300
description = 'Most of AsyncAPI features support, ready-to-go protocol implementations, reference resolver, codegen process customization, encoders/decoders, etc.'
+++

# Features

- AsyncAPI >=3.0.0 support
- Support the majority of [AsyncAPI entities](#asyncapi-support)
  - Servers, channels, operations, messages, schemas, parameters, correlation IDs, etc. (see below)
  - JSONSchema
  - JSONSchema polymorphism (oneOf, anyOf, allOf)
  - Specification extensions (`x-` flags), that control the code generation process
- Support many [protocols](#protocols)
- Support several AsyncAPI documents at once
- YAML configuration file
- Building the no-code CLI client executable with basic send-receive functionality
- Generating the infrastructure-as-code (IaC) files in [supported formats](#infrastructure-as-code-iac-generation).
- Go code generation
  - Implementation-independent code
  - Plugged-in client [implementations]({{< relref "/protocols-and-implementations" >}}) based on most 
    popular libraries
    - At least one implementation for every supported protocol is available to use
    - Automatic injection only necessary implementations to the generated code
    - Can be disabled for a particular protocol or at all if you use your own library
  - "Consumer only", "producer only" mode
  - `sync.Pool`-friendly code
  - [Content types](#content-types) support
  - [Code layout customization]({{< relref "/features/code-layout" >}})
  - [Code reuse]({{< relref "/features/code-reuse" >}})
  - Discarding parts of the document from the generation
  - Automatic formatting the generated code by `gofmt` (can be disabled)
  - Automatic determining the user project's module name
- Internal references (`$ref`) support
- Customization via [templates]({{< relref "/features/templates" >}}) in [text/template](https://pkg.go.dev/text/template) format
  - Plenty of functions available. Provided by [github.com/go-sprout/sprout](https://github.com/go-sprout/sprout)
  - Templates are organized in a tree structure, allowing customization on any granularity level
  - Calls tracing, easing the pain of template debugging
- Following the [external references]({{< relref "/features/references" >}}) to other AsyncAPI documents (`$ref`)
  - Local files
  - Remote documents available via HTTP(S)
  - [Custom resolver]({{< relref "/features/references#custom-spec-resolver" >}}) -- a user-provided command that does the resolving
- Verbose logging in debug and trace levels

## Protocols

Here are the protocols that are supported by `go-asyncapi` for now:

- {{< figure src="images/amqp.svg" alt="AMQP" link="/protocols-and-implementations/amqp" class="brand-icon" >}} [AMQP]({{< relref "/protocols-and-implementations/amqp" >}})
- {{< figure src="images/kafka.svg" alt="Apache Kafka" link="/protocols-and-implementations/apache-kafka" class="brand-icon" >}} [Apache Kafka]({{< relref "/protocols-and-implementations/apache-kafka" >}})
- {{< figure src="images/http-small.png" alt="HTTP" link="/protocols-and-implementations/http" class="brand-icon" >}} [HTTP]({{< relref "/protocols-and-implementations/http" >}})
- {{< figure src="images/mqtt.svg" alt="MQTT" link="/protocols-and-implementations/mqtt" class="brand-icon" >}} [MQTT]({{< relref "/protocols-and-implementations/mqtt" >}})
- {{< figure src="images/ip.png" alt="IP" link="/protocols-and-implementations/ip" class="brand-icon" >}} [IP RAW sockets]({{< relref "/protocols-and-implementations/ip" >}})**&ast;**
- {{< figure src="images/redis.svg" alt="Redis" link="/protocols-and-implementations/redis" class="brand-icon" >}} [Redis]({{< relref "/protocols-and-implementations/redis" >}})
- {{< figure src="images/tcpudp.svg" alt="TCP" link="/protocols-and-implementations/tcp" class="brand-icon" >}} [TCP]({{< relref "/protocols-and-implementations/tcp" >}})**&ast;**
- {{< figure src="images/tcpudp.svg" alt="UDP" link="/protocols-and-implementations/udp" class="brand-icon" >}} [UDP]({{< relref "/protocols-and-implementations/udp" >}})**&ast;**
- {{< figure src="images/websocket.svg" alt="WebSocket" link="/protocols-and-implementations/websocket" class="brand-icon" >}} [WebSocket]({{< relref "/protocols-and-implementations/websocket" >}})

{{< hint info >}}
**&ast;** - not described in the AsyncAPI specification
{{< /hint >}}

## AsyncAPI support

The marked items below are supported by the `go-asyncapi` tool, unmarked items are in the roadmap.

For the reference, see [AsyncAPI specification](https://github.com/asyncapi/spec/blob/v3.0.0/spec/asyncapi.md)

AsyncAPI Entities:

- [x] AsyncAPI object
- [ ] Identifier object (`$id`)
- [ ] Info object
  - [ ] Contact object
  - [ ] License object
- [x] Default Content Type
- [x] Servers object
  - [x] Server object
    - [x] Server Variable object
    - [x] Server Bindings object
- [x] Channels object
  - [x] Channel object
    - [x] Channel Bindings object
- [x] Messages object
  - [x] Message object
    - [ ] Message Example object
    - [x] Message Bindings object
- [x] Operations object
  - [x] Operation object
    - [ ] Operation Trait object
    - [ ] Operation Reply object
      - [ ] Operation Reply Address object
    - [x] Operation Bindings object
- [x] Parameters object
  - [x] Parameter object
- [ ] Tags object
  - [ ] Tag object
- [ ] External Documentation object
- [x] Components object
- [ ] Security Scheme object
- [ ] OAuth Flows object
  - [ ] OAuth Flow object
- [x] Correlation ID object

Other features:

- [x] Reference object (`$ref`)
- [x] Schema object
    - [x] Primitive types (number, string, boolean, etc.)
    - [x] Object types
    - [x] Array types
    - [x] Polymorphism: OneOf, AnyOf, AllOf
- [x] Channel Address Expressions
- [ ] Multi Format Schema object (link to non-AsyncAPI document)
- [x] Runtime Expression
- [ ] Traits merge mechanism

## Content types

The following content types (MIME types) has the default implementation in the generated code:

- [JSON](https://pkg.go.dev/encoding/json) (application/json): `encoding/json`
- [YAML](https://gopkg.in/yaml.v3) (application/yaml, application/x-yaml, text/yaml, text/x-yaml, text/vnd.yaml): `gopkg.in/yaml.v3`

To add a new content type, see the [Adding content type](https://go-asyncapi.dev/articles/adding-content-type) article.

## Infrastructure as code (IaC) generation

The `go-asyncapi` tool supports the generation of the following files formats:

- [docker-compose](https://docs.docker.com/compose/)
