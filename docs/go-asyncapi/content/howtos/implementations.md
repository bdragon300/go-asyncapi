---
title: "Implementations"
weight: 730
description: "Working with Implementations"
---

# Implementations

Basically, Implementation is a wrapper around a popular library with the equal interface and semantics, which makes
the library suitable to use in the protocol-agnostic code, generated from the AsyncAPI document.

`go-asyncapi` supports many [protocols]({{< relref "/features#protocols" >}}), and has at least one default 
out-of-the-box Implementation for each one.

The sections below describe how to select the built-in Implementation or make your own.

## Built-in Implementations

### How to use

By default, `go-asyncapi` generates the default implementations for every protocol met in the provided AsyncAPI document.
Implementations list and what of them are default are available by `list-implementations` subcommand.

You can select a built-in Implementations in configuration file:

```yaml
implementations:
  - protocol: kafka
    name: sarama
```

### How to disable

Built-in Implementation for a particular protocol can be disabled in the configuration file:

```yaml
implementations:
  - protocol: kafka
    disable: true
```

To disable all built-in Implementations, use the `--disable-implementations` CLI flag or set this in the configuration file:

```yaml
code:
    disableImplementations: true
```

## Custom Implementation

### How to make

You can make your own Implementation. To send/receive messages or open/close connection/channel, the generated code 
uses the interfaces, described in `abstract.go` in "runtime" package for every protocol. See 
[github.com/bdragon300/go-asyncapi/run](https://github.com/bdragon300/go-asyncapi/tree/master/run)

Basically, you need the `Producer`/`Consumer`, that opens a connection to the server (if needed) and able to create
`Publisher`/`Subscriber`. Also, you need the envelope type, that represents a protocol-specific outgoing (`EnvelopeWriter`)
and incoming (`EnvelopeReader`) message.

See the "implementations" directory in `go-asyncapi` repository for examples.
Also, see [code structure]({{< relref "/code-generation/code-structure" >}}) for the details.

### Customizing the generated code

It's possible to additionally customize the generated code, that uses the implementation.
Use the [templating]({{< relref "/templating-guide/overview" >}}) for that. See templates `code/proto/<protocol>` in 
[template tree]({{< relref "/templating-guide/template-tree" >}}).

For example, let's add our code to `Seal*` methods in generated channels for "kafka" protocol:

```gotemplate
{{define "code/proto/kafka/channel/publishMethods/block1"}}
    {{goQual "fmt.Printf"}}("Seal{{. | goIDUpper}} for envelope: %v", envelope)
{{end}}
```
