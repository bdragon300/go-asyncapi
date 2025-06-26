---
title: "Correlation ID"
weight: 310
description: "Correlation ID implementation notes" 
---

# Correlation ID

The Correlation ID is typically used to track the flow of messages and to ensure that messages are processed in the 
correct order. For more information, see the [AsyncAPI specification](https://www.asyncapi.com/docs/reference/specification/v3.0.0#correlation-id-object).

`go-asyncapi` supports the Correlation ID and generates the corresponding Go code for setting and getting its value
in the generated message structs.

## Overview

The Correlation ID contains the path where this id is stored in the message. For example:

```yaml
correlationId:
  location: $message.payload#/field1/10/field2
```

It can point either to payload or header. After the `#` symbol, it contains a path to the field.

## Special symbols encoding

AsyncAPI specification states that Correlation id location contains the
[JSON Pointer](https://datatracker.ietf.org/doc/html/rfc6901) after the `#` symbol. According to the specification,
the JSON Pointer is a string that contains a sequence of encoded symbols separated by `/`. 

Special symbols are *all symbols except the alphanumeric characters, `-`, `.` and `_`*.
So, they must be encoded according to the rules described below.

* Tilda symbol `~` must be written as `~0`
* Forward slash `/` must be written as `~1`
* Other symbols must be percent-encoded as described in [RFC 3986](https://tools.ietf.org/html/rfc3986#section-2.1)
  using the `%` character followed by two hexadecimal digits ([encoding table](https://www.w3schools.com/tags/ref_urlencode.ASP))
* Path items wrapped in quotes (single or double) are always treated as strings. Quotes are stripped before path evaluation.

For example, the path with three parts `~field _1`, `10` (must be a string json key, not an integer array index) and `"field2"/foo` 
must be written as `$message.payload#/~0field%20_1/'10'/%22field2%22~1foo`.
