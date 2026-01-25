---
title: "Overview"
weight: 910
description: "Overview of using the Go templates to customize the result produced by go-asyncapi"
---

# Overview

`go-asyncapi` uses the [Go templates](https://pkg.go.dev/text/template) to generate the code, client application, IaC files, etc.

Template files have the `.tmpl` extension, they are loaded on `go-asyncapi` start.  

The templates are organized in a [tree structure]({{< relref "/templating-guide/template-tree" >}}),
giving the user a way to customize the final generation result on any granularity level.

During the rendering phase, `go-asyncapi` executes only the **root template** (`main.tmpl` by default). 
All other templates in tree are [called inside it recursively](https://pkg.go.dev/text/template#hdr-Nested_template_definitions).
For several commands (`code`, `client`, etc.), `go-asyncapi` calls the **preamble template** (`preamble.tmpl` by default)
at the end for every generated Go source code. Typically, it adds the package declaration, import statements, 
"copyright" notice to the beginning of the file.

Templates has access to a bunch of functions that can be used in the template code.
Besides the [standard functions](https://pkg.go.dev/text/template#hdr-Functions), there are the
[sprout](https://docs.atom.codes/sprout/registries/list-of-all-registries) functions and 
[tool-specific]({{< relref "/templating-guide/functions" >}}) functions.

### Templates overriding

You can override any template in the tree by providing your own template file with the same name.

For example, let's add a method `Name()` to the server interface in channel code, by overriding the 
`code/proto/channel/serverInterface` template.
Put this to any file, say `my_templates/long_integer.tmpl`:

```gotemplate
{{define "code/proto/channel/serverInterface"}}
type {{ .Channel | goID }}Server{{.Protocol | goID}} interface {
    Open{{.Channel | goID}}{{.Protocol | goID}}({{goPkgExt "context"}}Context{{if .Parameters.Len}},{{goID .Channel}}Parameters{{end}}{{if .BoundOperations}},{{goPkgRun}}AnySecurityScheme{{end}}) (*{{. | goID}}{{.Protocol | goID}}, error)
    {{if .IsPublisher}}Producer() {{goPkgUtil .Protocol}}Producer{{end}}
    {{if .IsSubscriber}}Consumer() {{goPkgUtil .Protocol}}Consumer{{end}}
    Name() string
}
{{- end}}
```

Then specify the template directory when generating the code:

```bash
go-asyncapi code --template-dir ./my_templates my_asyncapi.yml
```

{{% hint info %}}
Overriding the whole template files works the same way, just name your template file as the original template name.
E.g. `channel.tmpl`.
{{% /hint %}}

## Concepts

### Template context

The context is object that keeps all information available for the template code, such as the rendered artifact, 
configuration options, current package name, current layout rule, etc.
It is passed to the root template in `data` argument of the [Execute](https://pkg.go.dev/text/template#Template.Execute) 
method and accessed by `.` and `$` operators.
Context object does not survive between the root template executions.

{{% hint info %}}
Context types are defined in [tmpl](https://github.com/bdragon300/go-asyncapi/blob/master/internal/tmpl/context.go) package.
{{% /hint %}}

### Definition and usage code

Artifacts representing Go types (i.e. ones that satisfy `common.GolangType` interface producing from JSONSchema object) 
may be rendered in the generated code in two forms: as a Go type definition and as a code snippet consuming this type
from the package where it was defined. The definition code is rendered by **definition** template, the type 
consuming code is rendered by **usage** template.

Every Go type artifact has two such templates. For example, for `lang.GoStruct` it will be 
`code/lang/gostruct/definition` and `code/lang/gostruct/usage`. Full template list is available 
[here]({{< relref "/templating-guide/template-tree" >}})

For example, we have the `lang.GoStruct` with a couple of fields.
`{{ goDef . }}` function called in on "foo" package produces the **definition** code:

```go
type MyStruct struct {
    Field1 string `json:"field1"`
    Field2 string `json:"field2"`
}
```

The `var x {{ goUsage . }}` in "bar" package produces the **usage** code (import from `foo` is added automatically):

```go
import foo

//...

var x foo.MyStruct
```

### Pinning

`go-asyncapi` can generate the code in any [layout]({{< relref "/howtos/customize-the-code-layout" >}}) you want. 
This feature gives much flexibility, but because of this we can't just hardcode all imports of the generated packages 
in templates, since we don't know the code layout prior the generation time.

To manage this, template code must manually "associate" an artifact with the current package name 
in runtime -- process called **pinning**. Once pinned (to the current rendering file), the artifact's location becomes 
known, and now it can be imported anywhere in the generated code.

So, whatever layout the user chose, the generated Go code will be correct.

{{% hint default %}}
For example, while rendering the `MyChannel` we should pin it the current file `channels/my_channel.go` by calling a function 
`{{pin .}}`. Much further, while rendering the `MyOperation`, to use this channel we can call `{{goPkg .Channel}}MyChannel`, that
automatically adds a correct import and produces the correct code `channels.MyChannel`.

The thing here is even user sets a code layout other than default, say, when channels and operations are placed
in the same package, the generated Go remains correct: no import will be produced in `MyOperation` code, and the same 
expression `{{goPkg .Channel}}MyChannel` will produce just `MyChannel`.
{{% /hint %}}

{{% hint warning %}}
Only the artifacts that *can be imported from somewhere* are pinnable. All other objects are not pinnable, including 
artifacts representing the Go primitive types (`int`, `byte`, etc.). 
Template function `pin` returns error for non-pinnable arguments.
{{% /hint %}}
