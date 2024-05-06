+++
title = 'Features'
bookCollapseSection = true
weight = 300
description = 'Most of AsyncAPI features support, ready-to-go protocol implementations, reference resolver, codegen process customization, encoders/decoders, etc.'
+++

# Features

- AsyncAPI 2.6.0
- Support the majority of [AsyncAPI features](#asyncapi-entities)
- Support many [protocols](#protocols)
- No extra dependencies in the generated code
- Optional simple client [implementations]({{< relref "/protocols-and-implementations" >}}) based on most 
  popular libraries
- [Reuse the code]({{< relref "/features/code-reuse" >}}) generated before
- [Breaking down the generated code]({{< relref "/features/code-breakdown" >}}) in several ways
- [Objects selection]({{< relref "/features/code-selection" >}}) to generate
- "Consumer only", "producer only" code generation
- [Content types](#content-types) support
- [References ($ref) resolving]({{< relref "/features/references" >}})
    - Document-local refs
    - Refs to files on local filesystem
    - Refs to the remote documents available via HTTP(S)
    - [Custom resolver]({{< relref "/features/references#custom-spec-resolver" >}}) (just an executable you provide), if refs are needed to be resolved in a custom way
- Optional encoders/decoders for content types, specified in the AsyncAPI document
- Support many features of jsonschema, including polymorphism (oneOf, anyOf, allOf)
- Support the zero-allocation approach if you need to reduce the load on the Go's garbage collector


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

{{< hint warning >}}
**&ast;** - not described in the AsyncAPI specification
{{< /hint >}}

## AsyncAPI entities

The marked items below are supported by the `go-asyncapi` tool, unmarked items are in the roadmap.

For reference see AsyncAPI specification https://github.com/asyncapi/spec/blob/v2.6.0/spec/asyncapi.md

- [x] AsyncAPI object **&ast;&ast;**
    - [x] Default Content Type
    - [ ] Identifier
- [ ] Info object
- [ ] Contact object
- [ ] License object
- [x] Servers object
    - [x] Server object
        - [x] Server Variable object
        - [x] Server Bindings object
- [x] Channels object
    - [x] Channel item object
        - [x] Operation object
            - [ ] Operation Trait object
            - [x] Operation Bindings object
        - [x] Channel Bindings object
- [x] Message object **&ast;&ast;**
    - [ ] Message Trait object
    - [ ] Message Example object
    - [x] Message Bindings object
- [ ] Tags object
    - [ ] Tag object
- [ ] External Documentation object
- [x] Components object
- [x] Reference object
- [x] Schema object **&ast;&ast;**
    - [x] _Primitive types (number, string, boolean, etc.)_
    - [x] _Object types_
    - [x] _Array types_
    - [x] _OneOf, AnyOf, AllOf_
- [ ] Security Scheme object
- [ ] Security Requirement object
- [ ] OAuth Flows object
    - [ ] OAuth Flow object
- [x] Parameters object
    - [x] Parameter object
- [x] Correlation ID object

{{< hint info >}}
**&ast;&ast;** - partial support, some entity fields are not supported yet
{{< /hint >}}

## Content types

For the following content types the `go-asyncapi` generates the default encoders/decoders. You can freely add your 
own content type with appropriate encoders/decoders or replace the default ones.

- [x] [JSON](https://pkg.go.dev/encoding/json) (application/json)
- [x] [YAML](https://gopkg.in/yaml.v3) (application/yaml, application/x-yaml, text/yaml, text/x-yaml, text/vnd.yaml)
