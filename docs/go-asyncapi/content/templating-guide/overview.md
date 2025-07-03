---
title: "Overview"
weight: 910
description: "Overview of using the Go templates to customize the result produced by go-asyncapi"
---

# Templating guide

`go-asyncapi` uses the [Go templates](https://pkg.go.dev/text/template) to generate the code, client application, IaC files, etc.

To customize the result, put all template you want to use to the directory and specify this directory
with the `--template-dir` flag or set an appropriate option in config.

## Structure

The templates are organized in a [tree structure]({{< relref "/templating-guide/template-tree" >}}),
that allows to customize the result on any granularity level. 

The whole tree is rendered starting from the **root template**, called `main.tmpl` by default.

The **preamble template** is executed at the end, after all root template invocations. 
Preamble template is called once per file, producing the output that is substituted at the beginning a file. Typically,
it contains the package declaration, import statements, "copyright" notice, etc. By default, its name `preamble.tmpl`.

## Naming

Templates files must have the `.tmpl` extension.

The file name is used as a template name if it doesn't contain any
[nested templates](https://pkg.go.dev/text/template#hdr-Nested_template_definitions) (`define` directives).
Otherwise, every nested templates are available by their names and the file name is ignored.

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
For this, `go-asyncapi` executes the root template separately for every artifact with `.Selected == true` and 
merges the results into a files and packages according the [code layout]({{< relref "/howtos/code-layout" >}}).

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

### Namespace

Namespace stores all the names and artifacts defined in templates between root templates executions.

The main its purpose is conditional rendering to avoid the name collisions in corner cases. 
For example, when an entity may be referenced from several places in document, its definition will be rendered several 
times, which is semantic error in Go code.

We have `goDef`, `def` functions to define the artifact/name and `defined`, `ndefined` functions
to check if the artifact/name is already defined in the current namespace. If you familiar with C/C++ languages, 
the namespace behavior may remind how `#define`, `#ifdef`, `#ifndef` preprocessor directives work.

For more information, see the [functions reference]({{< relref "/templating-guide/functions#defined" >}}).

## Usage

For example, we want to add additional prefix `My` to the name of the generated server interface generated near with every 
`channel`.

For that, create a file with any name, say `my_server_interface.tmpl`, copy the `code/proto/channel/serverInterface`
template from the 
[proto_channel.tmpl](https://github.com/bdragon300/go-asyncapi/blob/master/templates/code/proto/proto_channel.tmpl)
and modify it as follows:

```gotemplate
{{define "code/proto/channel/serverInterface"}}
type My{{ .Channel | goIDUpper }}Server{{.Protocol | goIDUpper}} interface {
    Open{{.Channel | goIDUpper}}{{.Protocol | goIDUpper}}(ctx {{goQual "context.Context"}}, {{if .ParametersType}}params {{ .ParametersType | goUsage }}{{end}}) (*{{ .Type | goUsage }}, error)
    {{if .IsPublisher}}Producer() {{goQualR .Protocol "Producer"}}{{end}}
    {{if .IsSubscriber}}Consumer() {{goQualR .Protocol "Consumer"}}{{end}}
}
{{- end}}
```

Now, run the `go-asyncapi` tool with the `--template-dir` flag pointing to the directory with the `my_server_interface.tmpl` file
and your version of `code/proto/channel/serverInterface` template will replace the default one:

```bash
go-asyncapi code --template-dir ./my_templates my_asyncapi.yml
```

{{% hint info %}}
Overriding the template files without ff works the same way, but you need to name your template file as the original template name.
E.g. `channel.tmpl`.
{{% /hint %}}
