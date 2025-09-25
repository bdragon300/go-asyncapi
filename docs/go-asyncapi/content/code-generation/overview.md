---
title: "Overview"
weight: 410
description: "Overview of the code generation capabilities of go-asyncapi" 
---

# Code generation overview

One of core features of `go-asyncapi` is Go boilerplate code generation from AsyncAPI documents.
The generated code has minimal dependencies on external libraries and contains the basic logic sufficient to
send and receive messages. You also can plug in the protocol implementations built-in in `go-asyncapi`, they are based on
popular libraries for that protocol. Also, it is possible to import the third-party code in the code being generated.

Code generation is also a crucial part of 
[Contract-First Development](https://www.moesif.com/blog/technical/api-development/Mastering-Contract-First-API-Development-Key-Strategies-and-Benefits/)
(so-called "API First" or "Schema First").

## Quick start

`go-asyncapi` accepts an AsyncAPI document in YAML or JSON format and produces the Go code. Repeated tool runs produce
exactly the same code.

Download the [example document](https://github.com/asyncapi/spec/blob/master/examples/streetlights-mqtt-asyncapi.yml) 
and save it as `streetlights-mqtt-asyncapi.yml`. Then, run the following command to generate the code, passing the document:

```bash
go-asyncapi code streetlights-mqtt-asyncapi.yml
```

Alternatively, fetching by URL is also supported:

```bash
go-asyncapi code --allow-remote-refs https://raw.githubusercontent.com/asyncapi/spec/refs/heads/master/examples/streetlights-mqtt-asyncapi.yml
```

{{% details title="Console output" %}}
```console
INFO Logging to stderr INFO
INFO Compilation 🔨: Compile a document url=streetlights-mqtt-asyncapi.yml
INFO Locating 📡: Reading document from filesystem path=streetlights-mqtt-asyncapi.yml
INFO Linking 🔗: Linking complete refs=22
INFO Rendering 🎨: Objects rendered count=0
INFO Formatting 📐: Formatting complete files=23
INFO Writing 📝: Writing complete files=23
INFO Code generation finished
INFO Done
```
{{% /details %}}

### Target directory

By default, the code is put to the `./asyncapi` directory (target directory) according to 
[default code layout]({{<relref "howtos/code-layout.md#default-layout">}}).

You can use the `-t` option to specify a different target directory:

```bash
go-asyncapi code ./my-asyncapi-code streetlights-mqtt-asyncapi.yml -t /tmp/my-asyncapi-code
```

### Debugging

Debug logging output is enabled by the `-v=1`. The `-v=2` enables the trace logging:

```bash
go-asyncapi code streetlights-mqtt-asyncapi.yml -v=1
```