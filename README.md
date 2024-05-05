# go-asyncapi

![GitHub go.mod Go version (subdirectory of monorepo)](https://img.shields.io/github/go-mod/go-version/bdragon300/go-asynapi)
![GitHub Workflow Status (with branch)](https://img.shields.io/github/actions/workflow/status/bdragon300/go-asyncapi/go-test.yml?branch=master)

[Documentation](https://bdragon300.github.io/go-asyncapi/)

`go-asyncapi` is the codegen tool that generates the boilerplate Go code from [AsyncAPI](https://www.asyncapi.com/)
documents.
It supports most of the AsyncAPI features, such as messages, channels, servers, bindings, correlation ids, etc.

> **[AsyncAPI](https://www.asyncapi.com/)** is a specification for defining APIs for event-driven architectures. The
> AsyncAPI document describes the messages, channels, servers, and other entities that the systems in event-driven
> architecture use to communicate with each other.

The generated code is not only just a bunch of stubs, it contains the abstract logic sufficient to send and
receive data through channels with no external dependencies except the standard Go library.
So, no extra features are inside beyond what is necessary (such as logging, metrics, etc.) --
it's up to you what you use in your project.

The code is also modular, so many generated objects can be used separately or be reused from another location.

Finally, `go-asyncapi` provides a pluggable **implementation** for every supported protocol — minimal client code
based on one of popular libraries for that protocol. This is convenient for simple needs or may be used as quickstart
for your own implementation.

Full list of features available on [Features](https://bdragon300.github.io/go-asyncapi/docs/features) page.

*Batteries included, but removable* :)

## Installation

```bash
go install github.com/bdragon300/go-asyncapi/cmd/go-asyncapi@latest
```

## Project versioning

`go-asyncapi` uses [Semantic Versioning](https://semver.org/) for versioning. For example, `1.4.0`.

Releasing a *patch version* contains only bug fixes and minor improvements. You won't need to regenerate the code after
upgrading the tool. E.g. **1.4.0 &rarr; 1.4.1**.

Releasing a *minor version* means that the generated code may be affected, but without breaking changes. You may need to
regenerate the code. E.g. **1.4.0 &rarr; 1.5.0**.

*Major version* release typically introduces the breaking changes. You may need to regenerate the code, to fix your
projects that uses it or to change the tool command line. E.g. **1.4.0 &rarr; 2.0.0**.

## FAQ

### Why do I need another third-party codegen tool? We have already the [official generator](https://github.com/asyncapi/generator)

**Long story short**: this one provides more features and protocols, and it has written in Go.

The official generator is quite specific for many cases.
At the moment, it produces the Go code bound with the Watermill framework, but not everyone uses the Watermill in
their projects.
Also, it supports only the AMQP protocol.

Instead, `go-asyncapi`:

* produces framework-agnostic code with the standard Go library as single dependency.
* supports more
  [protocols](https://bdragon300.github.io/go-asyncapi/docs/features#protocols) and more specific AsyncAPI entities, such as
  bindings, correlation ids, server variables, etc.
* contains the pluggable minimal clients for all supported protocols based on popular libraries.
* written in Go, so no need to have node.js or Docker or similar tools to run the generator.

*Another reason is that I don't know JavaScript well. And I'm not sure that if we want to support all AsyncAPI features,
the existing template would not be rewritten from the ground.*

### How can I customize the generated code templates?

`go-asyncapi` has many ways to customize the generated code, see the command line flags and `x-` fields description.

However, unlike the more common approach for other codegen tools that use templates, the `go-asyncapi` uses the
[jennifer](https://github.com/dave/jennifer) library. This approach is less customizable by user, but more
flexible. It's easier to make support of AsyncAPI specification more complete this way and to deal with complex
documents with plenty of interlinked objects.

But still, user template support in some reduced form is a planned feature.

## Alternatives

* https://github.com/asyncapi/generator (official generator)
* https://github.com/lerenn/asyncapi-codegen
* https://github.com/c0olix/asyncApiCodeGen
