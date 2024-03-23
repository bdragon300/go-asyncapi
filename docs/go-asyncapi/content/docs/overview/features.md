+++
title = 'Features'
weight = 10
+++

# Features

## Codegen tool
#TODO: reorder

- AsyncAPI documents in JSON and YAML formats are supported
- No extra dependencies in the generated code
- References ($ref) resolver
  - Document-local refs
  - External refs to files on the local filesystem
  - External refs to the remote documents (fetching them via HTTP/HTTPS)
  - You can provide your own resolver (an executable), if refs are needed to be resolved in a custom way
- Generated code customization:
  - Source code files and packages scoping
  - Excluding the particular objects or the whole document sections from the generation
  - Specification extensions (`x-` fields) support to control the generation result for particular objects
  - Generating only the consume code or only the produce code (or both)
- Importing a part of the source code generated before by `go-asyncapi`
- Pluggable minimal [implementations]({{< relref "/docs/protocols-and-implementations" >}}) for supported protocols. 
  You can also refuse to include them in your project and use your own implementation.
- Pluggable encoders/decoders for message content types. You can use you own encoder/decoder for any other content 
  type or replace the default one.
- Support many features of jsonschema, including polymorphism (oneOf, anyOf, allOf)
- Support the zero-allocation approach if you need to reduce the load on the Go's garbage collector

## Protocols

Here are the protocols that are supported by `go-asyncapi` for now:

- [x] [Kafka](https://kafka.apache.org/)
- [x] [AMQP](https://www.amqp.org/)
- [x] [MQTT](https://mqtt.org/)
- [x] [WebSockets](https://tools.ietf.org/html/rfc6455)
- [x] [Redis](https://redis.io/)
- [x] [HTTP](https://tools.ietf.org/html/rfc7230)
- [x] [TCP](https://tools.ietf.org/html/rfc793)
- [x] [UDP](https://tools.ietf.org/html/rfc768)
- [x] [Raw IP sockets](https://tools.ietf.org/html/rfc791)

## AsyncAPI entities

Below is the full list of AsyncAPI entities and which of these are supported by `go-asyncapi`:

- [x] AsyncAPI object (partially)
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
- [x] Message object (partially)
  - [ ] Message Trait object
  - [ ] Message Example object
  - [x] Message Bindings object
- [ ] Tags object
  - [ ] Tag object
- [ ] External Documentation object
- [x] Components object
- [x] Reference object
- [x] Schema object (partially)
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


## Default Encoders/Decoders for content types

- [x] JSON (application/json)
- [x] YAML (application/yaml, application/x-yaml, text/yaml, text/x-yaml, text/vnd.yaml)
