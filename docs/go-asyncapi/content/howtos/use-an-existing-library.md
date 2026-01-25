---
title: "Use an existing library"
weight: 550
description: "How to use an existing implementation library with the generated code"
---

# Use an existing library

Basically, the Implementation is a wrapper around a 3rd-party popular library with the standard interface and semantics, that 
just satisfy several interfaces.

`go-asyncapi` has many built-in [protocol implementations]({{< relref "/features#protocols" >}}). If `go-asyncapi`
does not have the implementation for a protocol, then it generates only the abstract code (interfaces, types, etc.) for this protocol.

There are two ways to use the existing code with the generated code:

1. Make types to satisfy the standard interfaces, pass it to channel/operation on creation and work with them as usual. See [example](TODO). 
2. Make types to satisfy the standard interfaces, convert the code to Go templates and include them into the code 
   generation process. See [example](TODO) and the description below.

## How implementation templates work

The implementation templates are processed in different way than the regular code templates. Because they don't depend
on AsyncAPI entities, every template is processed only once during the execution.

By the following configuration, we tell `go-asyncapi` to seek the implementation code templates for "foobar" protocol:

```yaml
code:
  implementation:
    custom:
      - protocol: "foobar"
        templateDirectory: "path/to/foobar/templates/directory"
```

Next, during the code generation, `go-asyncapi` reads every file with `*.tmpl` extension from template directory,
processes every template passing the `tmpl.CodeExtraTemplateContext` object and saves the result to the target
implementation directory with the same name replacing `.tmpl` extension with `.go`.

For example:

```
directory
├── foo.tmpl
└── subpackage
    └── bar.tmpl
```

produces the following files in the implementation directory (by default, `proto/foobar`):

```
target_dir
└── proto
    └── foobar
        ├── foo.go
        └── subpackage
            └── bar.go
```
