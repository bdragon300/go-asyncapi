---
title: "Overview"
weight: 910
description: "Overview of using the Go templates to customize the result produced by go-asyncapi"
---

# Templating guide

`go-asyncapi` uses the [Go templates](https://pkg.go.dev/text/template) to generate the code, client application, IaC files, etc.

## Naming

Templates files have the `.tmpl` extension.

Go Templates engine maintain its own template namespace, that is build during the parsing of template files tree. The
name in this namespace is used to identify and invoke the template inside other templates.

The namespace is built as follows:

* If the template file has any content, it is available by its file name.
* Each [nested template](https://pkg.go.dev/text/template#hdr-Nested_template_definitions)
  (i.e. `define` directive) becomes a separate template in the namespace, available by its name, regardless of the file 
  where it was defined.
* If file contains only nested templates, without any content outside of `define` directives, the file name is ignored.

### Structure

The templates are organized in a [tree structure]({{< relref "/templating-guide/template-tree" >}}),
that allows to customize the result on any granularity level.

The whole tree is rendered starting from the **root template**, called `main.tmpl` by default.

The **preamble template** (only for code and client generation) is executed at the end, after all root template invocations.
Preamble template is called once per file, producing the output that is substituted at the beginning a file. Typically,
it contains the package declaration, import statements, "copyright" notice, etc. By default, its name `preamble.tmpl`.

### Overriding templates

You can override any template in the tree by providing your own template file with the same name.

For example, we want to add a method `Name()` to the server interface in channel code. 
To do that, we need to override the `code/proto/channel/serverInterface` template. 
Let's create a file with any name, say `my_server_interface.tmpl` with the following content:

```gotemplate
{{define "code/proto/channel/serverInterface"}}
type {{ .Channel | goIDUpper }}Server{{.Protocol | goIDUpper}} interface {
    Open{{.Channel | goIDUpper}}{{.Protocol | goIDUpper}}(ctx {{goQual "context.Context"}}, {{if .ParametersType}}params {{ .ParametersType | goUsage }}{{end}}) (*{{. | goIDUpper}}{{.Protocol | goIDUpper}}, error)
    {{if .IsPublisher}}Producer() {{goQualR .Protocol "Producer"}}{{end}}
    {{if .IsSubscriber}}Consumer() {{goQualR .Protocol "Consumer"}}{{end}}
    Name() string
}
{{- end}}
```

Put this file in a directory, say `my_templates`, and run the `go-asyncapi` tool with the `--template-dir` flag pointing to this directory:
```bash
go-asyncapi code --template-dir ./my_templates my_asyncapi.yml
```

`go-asyncapi` will scan the `my_templates` directory and override the `code/proto/channel/serverInterface` template with our version.

{{% hint info %}}
Overriding the whole template files works the same way, just name your template file as the original template name.
E.g. `channel.tmpl`.
{{% /hint %}}

## Concepts

### Template context

The context object is passed to the root template in `data` argument of the `Execute` method, so it initially
available by `.` and `$` operators. The context object does not survive between the template executions.

The context holds the rendered artifact, configuration options, current package name, current layout rule, etc. 
Various "Context" structs are defined in `tmpl` package.

### Artifacts

The main object to work with in the templates is **artifact** -- the intermediate representation 
object of an AsyncAPI entity in document. All artifacts are defined in `render` and `lang` packages and satisfy 
the `common.Artifact` interface.

A part of artifacts represent the complex entities (e.g. channel) and produce the complex code. 
Others are simpler and represent a simple Go type (e.g. jsonschema object), they additionally satisfy the 
`common.GolangType` interface.

The templates goal is to render the artifacts into the desired output type: Go code, IaC configuration, etc.
For this, `go-asyncapi` executes the root template separately for every artifact that is `.Selected == true` and 
merges the results into a files after executions and packages according the [code layout]({{< relref "/howtos/code-layout" >}}).

The artifacts, compiled from several AsyncAPI documents, get to the same list. However, every artifact keeps the document URL 
and the location where it was defined.

### Definition and usage code

One thing that is worth to mention is main difference between `goUsage` and `goDef` functions. The `goUsage`
function render the **usage code** of the artifact, i.e. code snippet to "consume" the Go identifier in another place.
The `goDef` function renders the **definition code** of the artifact, i.e. the Go code with type declaration.

For example, we have the `lang.GoStruct` with a couple of fields.
`{{ goDef . }}` function called in on "foo" package produces the **definition code**:

```go
type MyStruct struct {
    Field1 string `json:"field1"`
    Field2 string `json:"field2"`
}
```

The `{{ goUsage . }}` in "bar" package produces the **usage code** (import from `foo` will be added automatically):

```go
foo.MyStruct
```

This also works for the artifacts defined in the current package (e.g. `MyStruct`), and in 
third-party packages (e.g. `mypackage.MyStruct` from "github.com/myuser/mypackage").

### Pinning

`go-asyncapi` can generate the code in any [layout]({{< relref "/howtos/code-layout" >}}) you want. 
This feature gives much flexibility, but the problem is that we can't hardcode imports in templates between the generated packages, 
because we don't know the code layout prior the generation time.

For example, we have a struct `MySchema` defined in the "schemas" package. The simplest way is just use `MySchema` in 
every template it uses, but this works only if `MySchema` is defined in the same package. 
For importing the "schemas" package from, say "servers" package, this won't work when the code layout is not default -- 
`MySchema` may be even put to the package other than "schemas".

To manage this, we use *pinning*. Pinning is the mechanism to "associate" the artifact with the current package name 
in runtime. Once pinned, the artifact's location becomes known, and now it can be imported anywhere from the generated code. 
In other words, somewhere in templates we pin the artifact by `pin` or `goDef` functions, and then somewhere else in templates
we can use `goPkg`, `goUsage` functions to import its package and use the artifact in the generated code.

In example above, we write `{{goDef .Type}}` to pin and draw the `MySchema` declaration, and in any place we use it 
we write `{{goPkg}}MySchema` which will render either `MySchema` or `schemas.MySchema` with a proper import regardless
the current code layout.

Some artifacts can not be pinned, because they can't belong to any package, e.g. primitive Go types like `string`, `int`, etc.