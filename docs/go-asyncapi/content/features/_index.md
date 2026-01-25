+++
title = 'Features'
weight = 100
bookToc = true
description = 'asyncapi-go features overview'
+++

# Features

- AsyncAPI 3.x specification
  - Majority of [AsyncAPI entities](#asyncapi-support)
  - [JSONSchema support](#jsonschema)
    - Extensible [object types and formats](#types-and-formats)
  - [Content types](#content-types)
  - [Automatic resolving]({{< relref "/asyncapi-specification/references" >}}) the `$ref`s
    - Fetching files from HTTP, local or by executing the user shell command
  - Specification extensions (`x-` fields), that control the code generation process
  - [Security schemes](#security-schemes)
- Generating the [boilerplate code]({{< relref "/commands/code" >}})
  - Abstract code for any protocol, even if it not well-known
    - Additional support code for some [protocols](#protocols) sufficient for basic
      send-receive functionality
  - Flexible codegen process control: [excluding entities]({{< relref "/howtos/exclude-an-asyncapi-entity" >}}),
    [publish-only / subscribe-only generation]({{< relref "/howtos/pub-sub-only" >}}), etc.
  - Setting any structure of the generated code using [code layout]({{< relref "/howtos/customize-the-code-layout" >}})
  - `sync.Pool`-friendly code
- Generating the no-code [CLI application executable]({{< relref "/commands/client" >}}) with basic send-receive 
  functionality
- Generating the [server definitions]({{< relref "/commands/infra" >}}) for the
  [supported engines](#server-definitions-generation).
- Visualizing AsyncAPI documents in [diagrams]({{< relref "/commands/diagram" >}})
  - SVG, D2 formats
  - Channel-centric, server-centric and combined views
  - Plenty of [customization options]({{< relref "/commands/diagram#customization" >}})
  - [Themes]({{< relref "/commands/diagram#themes" >}}) support
- Documentation web UI
  - [Serving]({{< relref "/commands/ui#serving-the-ui" >}}) using the built-in web server launching in one command
  - Generating the static HTML
  - [Bundling]({{< relref "/commands/ui#bundling-the-assets" >}}) support
- Customization the output by [user templates]({{< relref "/templating-guide/overview" >}}) in [text/template](https://pkg.go.dev/text/template) format
- Configuring via YAML [configuration file]({{< relref "/configuration" >}})
- Verbose logging in debug and trace levels

{{% hint warning %}}
AsyncAPI 2.x is not supported
{{% /hint %}}

## Protocols

Here are the protocols that are supported by `go-asyncapi` for now:

- {{< figure src="images/amqp.svg" alt="AMQP" link="/protocols#amqp" class="brand-icon" >}} [AMQP]({{< relref "/protocols#amqp" >}})
- {{< figure src="images/http.svg" alt="HTTP" link="/protocols#http" class="brand-icon" >}} [HTTP]({{< relref "/protocols#http" >}})
- {{< figure src="images/ip.png" alt="IP" link="/protocols#ip-raw-sockets" class="brand-icon" >}} [IP RAW sockets]({{< relref "/protocols#ip-raw-sockets" >}})
- {{< figure src="images/kafka.svg" alt="Apache Kafka" link="/protocols#apache-kafka" class="brand-icon" >}} [Apache Kafka]({{< relref "/protocols#apache-kafka" >}})
- {{< figure src="images/mqtt.svg" alt="MQTT v3" link="/protocols#mqtt" class="brand-icon" >}} [MQTT v3]({{< relref "/protocols#mqtt-v3" >}})
- {{< figure src="images/mqtt.svg" alt="MQTT v5" link="/protocols#mqtt5" class="brand-icon" >}} [MQTT v5]({{< relref "/protocols#mqtt-v5" >}})
- {{< figure src="images/nats.svg" alt="NATS" link="/protocols#nats" class="brand-icon" >}} [NATS]({{< relref "/protocols#nats" >}})
- {{< figure src="images/redis.svg" alt="Redis" link="/protocols#redis" class="brand-icon" >}} [Redis]({{< relref "/protocols#redis" >}})
- {{< figure src="images/tcpudp.svg" alt="TCP" link="/protocols#tcp" class="brand-icon" >}} [TCP]({{< relref "/protocols#tcp" >}})
- {{< figure src="images/tcpudp.svg" alt="UDP" link="/protocols#udp" class="brand-icon" >}} [UDP]({{< relref "/protocols#udp" >}})
- {{< figure src="images/websocket.svg" alt="WebSocket" link="/protocols#websocket" class="brand-icon" >}} [WebSocket]({{< relref "/protocols#websocket" >}})

## AsyncAPI support

The marked items below are supported by the `go-asyncapi` tool (partially or fully), unmarked items are in the roadmap.

For the reference, see [AsyncAPI specification](https://github.com/asyncapi/spec/blob/v3.0.0/spec/asyncapi.md)

### AsyncAPI Entities

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
- [x] Security Scheme object
- [ ] OAuth Flows object
  - [ ] OAuth Flow object
- [x] Correlation ID object

### Other

- [x] [Reference object](https://github.com/asyncapi/spec/blob/master/spec/asyncapi.md#reference-object) (`$ref`)
- [x] [Channel Address Expressions](https://github.com/asyncapi/spec/blob/master/spec/asyncapi.md#channel-address-expressions)
- [ ] [Multi Format Schema object](https://github.com/asyncapi/spec/blob/master/spec/asyncapi.md#multi-format-schema-object)
- [x] [Runtime Expression](https://github.com/asyncapi/spec/blob/master/spec/asyncapi.md#runtime-expression)
- [ ] [Traits merge mechanism](https://github.com/asyncapi/spec/blob/master/spec/asyncapi.md#traits-merge-mechanism)

## JSONSchema

AsyncAPI uses superset of [JSON Schema Specification Draft 07](https://json-schema.org/draft-07) for message payloads and other 
schema definitions. See [AsyncAPI spec](https://github.com/asyncapi/spec/blob/master/spec/asyncapi.md#schema-object).

The following JSONSchema features are supported by `go-asyncapi`:

- [x] `type`: [see below](#types-and-formats)
- [ ] `additionalItems`
- [x] `additionalProperties`
- [x] `allOf`
- [x] `anyOf`
- [ ] `const`
- [ ] `contains`
- [ ] `default`
- [ ] `definitions`
- [ ] `deprecated`
- [x] `description`
- [ ] `discriminator`
- [ ] `else`
- [ ] `enum`
- [ ] `examples`
- [ ] `exclusiveMaximum`
- [ ] `exclusiveMinimum`
- [ ] `externalDocs`
- [x] `format`: [see below](#types-and-formats)
- [ ] `if`
- [x] `items`
- [ ] `maxItems`
- [ ] `maxLength`
- [ ] `maxProperties`
- [ ] `maximum`
- [ ] `minItems`
- [ ] `minLength`
- [ ] `minProperties`
- [ ] `minimum`
- [ ] `multipleOf`
- [ ] `not`
- [ ] `oneOf`
- [ ] `pattern`
- [ ] `patternProperties`
- [x] `properties`
- [ ] `propertyNames`
- [ ] `readOnly`
- [x] `required`
- [ ] `then`
- [x] `title`
- [ ] `uniqueItems`

### Types and formats

{{% hint note %}}
JSONSchema formats are extensible by user, [see more]({{< relref "/howtos/add-a-jsonschema-format" >}}).
{{% /hint %}}

The following object types and formats are supported by `go-asyncapi` out-of-the-box:

- `object`: `struct`, `map[T1]T2`, etc., depending on the schema definition
- `array`: `[]T`, `[x]T`, depending on the schema definition
- `boolean`: `bool`
- `string`: `string`
    - `date-time`, `datetime`: [time.Time](https://pkg.go.dev/time#Time)
    - `ipv4`, `ipv6`: [net.IP](https://pkg.go.dev/net#IP)
    - `uuid`: [github.com/google/uuid](https://pkg.go.dev/github.com/google/uuid#UUID)
    - `binary`, `bytes`: `[]byte`
- `integer`: `int`
    - `int8`: `int8`
    - `int16`: `int16`
    - `int32`: `int32`
    - `int64`: `int64`
    - `uint`: `uint`
    - `uint8`: `uint8`
    - `uint16`: `uint16`
    - `uint32`: `uint32`
    - `uint64`: `uint64`
- `number`: `float64`
    - `float`: `float32`
    - `double`: `float64`
    - `decimal`: [github.com/shopspring/decimal](https://pkg.go.dev/github.com/shopspring/decimal#Decimal)
- `null`: `any`

## Content types

{{% hint note %}}
Content types are extensible by user, [see more]({{< relref "/howtos/add-a-content-type" >}}).
{{% /hint %}}

The following content types (MIME types) are supported by `go-asyncapi` out-of-the-box:

- `application/json`: [encoding/json](https://pkg.go.dev/encoding/json)
- `application/yaml`, `application/x-yaml`, `text/yaml`: [gopkg.in/yaml.v3](https://gopkg.in/yaml.v3)
- `application/binary`: [encoding/gob](https://pkg.go.dev/encoding/gob)
- `text/plain`: built-in conversion to/from string
- `application/xml`: [encoding/xml](https://pkg.go.dev/encoding/xml)

## Security schemes

{{% hint note %}}
Security schemes are extensible by user, [see more]({{< relref "/howtos/add-a-security-type" >}}).
{{% /hint %}}

The following security scheme types are supported by `go-asyncapi`:

* [x] `userPassword`
* [x] `apiKey`
* [ ] `X509`
* [ ] `symmetricEncryption`
* [ ] `asymmetricEncryption`
* [ ] `httpApiKey`
* [ ] `http`
* [ ] `oauth2`
* [ ] `openIdConnect`
* [ ] `plain`
* [ ] `scramSha256`
* [ ] `scramSha512`
* [ ] `gssapi`

## Server definitions generation

The `go-asyncapi` tool supports the generation for the following engines:

- [docker-compose](https://docs.docker.com/compose/)
