+++
title = 'Features'
weight = 100
description = 'asyncapi-go features overview'
+++

# Features

- AsyncAPI >=3.0.0 support (2.6.0 is not supported)
- Support the majority of [AsyncAPI entities](#asyncapi-support)
  - Servers, channels, operations, messages, schemas, parameters, correlation IDs, etc. (see below)
  - JSONSchema
  - JSONSchema polymorphism (oneOf, anyOf, allOf)
  - Specification extensions (`x-` fields), that control the code generation process
- Support many [protocols](#protocols)
- Support several AsyncAPI documents at once
- YAML [configuration file]({{< relref "/configuration" >}})
- Generating the no-code [CLI client executable]({{< relref "/client-application-generation" >}}) with basic send-receive 
  functionality
- Generating the [infrastructure-as-code (IaC) files]({{< relref "/infrastructure-files-generation" >}}) in 
  [supported engines](#infrastructure-as-code-iac-generation).
- Visualizing the AsyncAPI document as a [diagram]({{< relref "/visualization" >}})
  - SVG (default) and D2 formats supported
  - Channel-centric, server-centric and combined (default) views
  - Supports multiple documents at once, showing relationships between them
  - Plenty of [customization options]({{< relref "/visualization#customization" >}})
  - [Themes]({{< relref "/visualization#themes" >}}) support
  - D2 output is also customizable via [templates]({{< relref "/templating-guide/overview" >}})
- Go code generation
  - [Implementation-agnostic code]({{< relref "/code-generation/code-structure" >}})
  - Built-in ready to use [implementations]({{< relref "/protocols-and-implementations" >}}) based on most 
    popular libraries
    - At least one implementation for every [supported protocol](#protocols)
    - Automatic injection to the generated code (optional)
    - You can [make your own implementation]({{< relref "/howtos/implementations#how-to-make" >}}) and use it in the generated code
  - [Partial generation]({{< relref "/howtos/partial-generation" >}}) of the AsyncAPI document
    - [Pub/sub only]({{< relref "/howtos/partial-generation#pubsub-only" >}}) generation
    - Ignoring particular AsyncAPI entities in [document]({{< relref "/asyncapi-specification/special-fields#x-ignore" >}})
      and in [go-asyncapi config]({{< relref "/howtos/partial-generation#code-layout" >}})
  - [Content types](#content-types) support
    - [Adding a new]({{< relref "/howtos/content-types#adding-a-new-content-type" >}}) content type 
    - [Replacing]({{< relref "/howtos/content-types#replacing-the-default-encoderdecoder" >}}) the default encoder/decoder code
      for the supported content type
  - Flexible customization of [code layout]({{< relref "/howtos/code-layout" >}})
  - [Code reuse]({{< relref "/howtos/code-reuse" >}})
  - Automatic formatting by `gofmt`
  - Automatic determining the user project's module name
  - `sync.Pool`-friendly code
- Support of [internal references]({{< relref "/asyncapi-specification/references" >}}) (`$ref`) in document
- Support of [external references]({{< relref "/asyncapi-specification/references" >}}) (`$ref`) in document to other AsyncAPI documents
  - Locating a document in local files, by making HTTP(S) requests
  - For complex scenarios you can provide a [custom locator]({{< relref "/asyncapi-specification/references#custom-reference-locator" >}}) shell command
- Output customization via [templates]({{< relref "/templating-guide/overview" >}}) in [text/template](https://pkg.go.dev/text/template) format
  - Plenty of [functions]({{< relref "/templating-guide/functions" >}}) available. Provided by [github.com/go-sprout/sprout](https://github.com/go-sprout/sprout)
  - Templates are organized in a [tree structure]({{< relref "/templating-guide/template-tree" >}}), allowing customization on any granularity level
  - Tracing of function calls, easing the pain of template debugging
- Verbose logging in debug and trace levels

## Protocols

Here are the protocols that are supported by `go-asyncapi` for now:

- {{< figure src="images/amqp.svg" alt="AMQP" link="/protocols-and-implementations#amqp" class="brand-icon" >}} [AMQP]({{< relref "/protocols-and-implementations#amqp" >}})
- {{< figure src="images/kafka.svg" alt="Apache Kafka" link="/protocols-and-implementations#apache-kafka" class="brand-icon" >}} [Apache Kafka]({{< relref "/protocols-and-implementations#apache-kafka" >}})
- {{< figure src="images/http-small.png" alt="HTTP" link="/protocols-and-implementations#http" class="brand-icon" >}} [HTTP]({{< relref "/protocols-and-implementations#http" >}})
- {{< figure src="images/mqtt.svg" alt="MQTT" link="/protocols-and-implementations#mqtt" class="brand-icon" >}} [MQTT]({{< relref "/protocols-and-implementations#mqtt" >}})
- {{< figure src="images/ip.png" alt="IP" link="/protocols-and-implementations#ip-raw-sockets" class="brand-icon" >}} [IP RAW sockets]({{< relref "/protocols-and-implementations#ip-raw-sockets" >}})**&ast;**
- {{< figure src="images/redis.svg" alt="Redis" link="/protocols-and-implementations#redis" class="brand-icon" >}} [Redis]({{< relref "/protocols-and-implementations#redis" >}})
- {{< figure src="images/tcpudp.svg" alt="TCP" link="/protocols-and-implementations#tcp" class="brand-icon" >}} [TCP]({{< relref "/protocols-and-implementations#tcp" >}})**&ast;**
- {{< figure src="images/tcpudp.svg" alt="UDP" link="/protocols-and-implementations#udp" class="brand-icon" >}} [UDP]({{< relref "/protocols-and-implementations#udp" >}})**&ast;**
- {{< figure src="images/websocket.svg" alt="WebSocket" link="/protocols-and-implementations#websocket" class="brand-icon" >}} [WebSocket]({{< relref "/protocols-and-implementations#websocket" >}})
- {{< figure src="images/nats.svg" alt="NATS" link="/protocols-and-implementations#nats" class="brand-icon" >}} [NATS]({{< relref "/protocols-and-implementations#nats" >}})

{{< hint info >}}
&ast; - not described in the AsyncAPI specification
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
    - [x] Operation Reply object
      - [x] Operation Reply Address object
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
    - [x] Formats
    - [x] Object types
    - [x] Array types
    - [x] Polymorphism: OneOf, AnyOf, AllOf
- [x] Channel Address Expressions
- [ ] Multi Format Schema object (link to non-AsyncAPI document)
- [x] Runtime Expression
- [ ] Traits merge mechanism

## Content types

The following content types (MIME types) has the default implementation in the generated code:

- application/json: [encoding/json](https://pkg.go.dev/encoding/json)
- application/yaml, application/x-yaml, text/yaml: [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3)
- application/binary: [encoding/gob](https://pkg.go.dev/encoding/gob)
- text/plain: built-in conversion to/from string
- application/xml: [encoding/xml](https://pkg.go.dev/encoding/xml)

{{% hint tip %}}
You also can add any content type or override the existing one, [see more]({{< relref "/howtos/content-types#adding-a-new-content-type" >}}).
{{% /hint %}}

## JSONSchema formats

The formats and the produced Go types, supported by `go-asyncapi` out of the box:

- `string` type
  - `date-time`, `datetime`: [time.Time](https://pkg.go.dev/time#Time)
  - `ipv4`, `ipv6`: [net.IP](https://pkg.go.dev/net#IP)
  - `uuid`: [github.com/google/uuid](https://pkg.go.dev/github.com/google/uuid)
  - `binary`, `bytes`: `[]byte`
- `integer` type
  - `int8`: `int8`
  - `int16`: `int16`
  - `int32`: `int32`
  - `int64`: `int64`
  - `uint`: `uint`
  - `uint8`: `uint8`
  - `uint16`: `uint16`
  - `uint32`: `uint32`
  - `uint64`: `uint64`
- `number` type
  - `float`: `float32`
  - `double`: `float64`
  - `decimal`: [github.com/shopspring/decimal](https://pkg.go.dev/github.com/shopspring/decimal)

{{% hint tip %}}
You also can add any format or override the existing one, [see more]({{< relref "/howtos/jsonschema#adding-a-new-format" >}}).
{{% /hint %}}

## Infrastructure as code (IaC) generation

The `go-asyncapi` tool supports the generation for the following engines:

- [docker-compose](https://docs.docker.com/compose/)
