---
title: "Runtime expression"
weight: 730
description: "Runtime expression implementation notes" 
---

# Runtime expression

Runtime expression is expression with nested path to a specific value in the message payload or headers, 
that is resolved in runtime. See
[runtime expression](https://www.asyncapi.com/docs/reference/specification/v3.0.0#runtimeExpression).

Two main use cases of runtime expressions in AsyncAPI specification are **Correlation ID** and **Operation Reply Address**.

`go-asyncapi` supports the Runtime expression and generates the corresponding Go code for setting and getting its value
in the generated message structs.

## Correlation ID

The Correlation ID is typically used to track the flow of messages and to ensure that messages are processed in the 
correct order. For more information, see the [AsyncAPI specification](https://www.asyncapi.com/docs/reference/specification/v3.0.0#correlation-id-object).

For example, the following AsyncAPI snippet defines a Correlation ID that points to a field in the message payload:

```yaml
correlationId:
  location: $message.payload#/field1/10/field2
```

## Operation Reply Address

The Operation Reply Address denotes the field in message payload or headers that contains the address to which replies should be sent.

For example, the following AsyncAPI snippet defines an Operation Reply Address that points to a message header `replyTo`:

```yaml
operation:
  foo:
    action: receive
    channel: '$ref: #/channels/myChannel'
    reply:
      address: $message.headers#/replyTo
```


## Special symbols encoding

AsyncAPI specification states that Runtime expression contains the
[JSON Pointer](https://datatracker.ietf.org/doc/html/rfc6901) after the `#` symbol. According to the specification,
the JSON Pointer is a string that contains a sequence of encoded symbols separated by `/`. 

Special symbols are all symbols except the alphanumeric characters, `-`, `.` and `_`.
So, they must be encoded according to the rules described below.

* Tilda symbol `~` must be written as `~0`
* Forward slash `/` must be written as `~1`
* Other symbols must be percent-encoded as described in [RFC 3986](https://tools.ietf.org/html/rfc3986#section-2.1)
  using the `%` character followed by two hexadecimal digits ([encoding table](https://www.w3schools.com/tags/ref_urlencode.ASP))
* Path items wrapped in quotes (single or double) are always treated as strings. Quotes are stripped before path evaluation.

For example, the path with three parts `~field _1`, `10` (must be a string json key, not an integer array index) and `"field2"/foo` 
must be written as `$message.payload#/~0field%20_1/'10'/%22field2%22~1foo`.
