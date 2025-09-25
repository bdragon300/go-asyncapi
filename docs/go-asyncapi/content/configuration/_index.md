---
title: "Configuration reference"
weight: 800
bookToC: true
description: "go-asyncapi configuration reference"
---

# Configuration reference

Configuration is set in a single YAML file, that may contain the attributes described below. 
All attributes are optional, and if not set, the tool uses the default values.

{{% hint tip %}}
By default, `go-asyncapi` reads the configuration file named `go-asyncapi.yaml` or `go-asyncapi.yml` in the 
current working directory if it exists. You may set you own location by passing the `-c` CLI option.
{{% /hint %}}

{{% hint info %}}
Configuration file with all default values is 
[here](https://github.com/bdragon300/go-asyncapi/blob/master/assets/default_config.yaml)
{{% /hint %}}

## Config

| Attribute       | Type                                | Default                                                               | Description                                                                                              |
|-----------------|-------------------------------------|-----------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------|
| configVersion   | int                                 | `1`                                                                   | Configuration version. Now only the `1` is valid value                                                   |
| projectModule   | string                              |                                                                       | Project module name for the generated code. If empty, takes from go.mod in the current working directory |
| runtimeModule   | string                              | `github.com/bdragon300/go-asyncapi/run`                               | Path to runtime module with auxiliary code                                                               |
| layout          | [][Layout](#layout)                 | [Default layout]({{< relref "/howtos/code-layout#default-layout" >}}) | Generated code layout rules                                                                              |
| locator         | [Locator](#locator)                 |                                                                       | [Reference locator]({{< relref "/asyncapi-specification/references#reference-locator" >}}) settings      |
| implementations | [][Implementation](#implementation) |                                                                       | Per-protocol implementations generation settings                                                         |
| code            | [Code](#code)                       |                                                                       | Code generation settings                                                                                 |
| client          | [Client](#client)                   |                                                                       | No-code client building settings                                                                         |
| infra           | [Infra](#infra)                     |                                                                       | Infra files generation settings                                                                          |
| diagram         | [Diagram](#diagram)                 |                                                                       | Diagram generation settings                                                                              |

## Layout

| Attribute        | Type                             | Default | Description                                                                                 |
|------------------|----------------------------------|---------|---------------------------------------------------------------------------------------------|
| nameRe           | string                           |         | Condition: regex to match the entity name                                                   |
| artifactKinds    | []string                         |         | Condition: match to one of [artifact kinds](#artifact-kind) in list                         |
| moduleURLRe      | string                           |         | Condition: regex to match the document URL or path                                          |
| pathRe           | string                           |         | Condition: regex to match the artifact path inside a document (e.g. `#/channels/myChannel`) |
| protocols        | []string                         |         | Condition: match to any [protocol](#protocols-and-implementations) in list                  |
| not              | bool                             | `false` | If `true`, apply NOT operations to the match                                                |
| render           | [LayoutRender](#layoutrender)    |         | Render settings                                                                             |
| reusePackagePath | string                           |         | Path to the Go package to [reuse]({{< relref "/howtos/code-reuse" >}}) the code from.       |

Right before the rendering stage, `go-asyncapi` compiles the "artifacts" (intermediate representation objects) 
from the AsyncAPI document. To determine how and where their generated code should be rendered to, it checks every 
artifact one-by-one against all rules in the layout.

If the artifact matches a rule, `go-asyncapi` generates its code and writes it to the file from the matched rule with given settings.
After that it goes to the next rule, and so on, until all rules are checked.

If the artifact does not match any rule, it is skipped and not rendered.

{{% hint info %}}
Several conditions in one layout rule are joined via **AND** operation.
{{% /hint %}}

{{% hint info %}}
For more examples of the layout rules, see the [Code layout]({{< relref "/howtos/code-layout" >}}) page.
{{% /hint %}}

## LayoutRender

| Attribute        | Type       | Default                 | Description                                                                                                         |
|------------------|------------|-------------------------|---------------------------------------------------------------------------------------------------------------------|
| file             | string     |                         | **Required**. [Template expression](#template-expressions) with output file name. Dot is `tmpl.CodeTemplateContext` |
| package          | string     | Directory name          | Package name if it differs from the directory name                                                                  |
| protocols        | []string   | All supported protocols | [Protocols](#protocols-and-implementations) that are allowed to render                                              |
| protoObjectsOnly | bool       | `false`                 | If `true`, render only the protocol-specific entity code (e.g. `Channel1Kafka`, but skip `Channel1`)                |
| template         | string     | `main.tmpl`             | Root template name, used for rendering.                                                                             |

## Locator

| Attribute             | Type          | Default | Description                                                                                                                                           |
|-----------------------|---------------|---------|-------------------------------------------------------------------------------------------------------------------------------------------------------|
| allowRemoteReferences | bool          | `false` | `true` allows reading remote `$ref`s (denied by default by security reasons)                                                                          |
| rootDirectory         | string        |         | The base directory to locate the files. Not used if empty.                                                                                            |
| timeout               | time.Duration | `30s`   | Reference resolving timeout (sets both the HTTP request timeout and command execution timeout)                                                        |
| command               | string        |         | Shell command of [custom locator]({{< relref "/asyncapi-specification/references#custom-reference-locator" >}}). If empty, uses the built-in locator. |

{{% hint info %}}
For more information about the reference locator, see the [References]({{< relref "/asyncapi-specification/references" >}}) page.
{{% /hint %}}

## Implementation

| Attribute        | Type       | Default | Description                                                                                                                                                                                                |
|------------------|------------|---------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| protocol         | string     |         | [Protocol](#protocols-and-implementations)                                                                                                                                                                 |
| name             | string     |         | Implementation name.                                                                                                                                                                                       |
| disable          | bool       | `false` | If `true`, disables the implementation code generation.                                                                                                                                                    |
| directory        | string     |         | [Template expression](#template-expressions) with the output directory name, relative to the target directory. By default, takes from `code.implementationsDir` setting. Dot is `tmpl.ImplTemplateContext` |
| package          | string     |         | Package name if it differs from the directory name                                                                                                                                                         |
| reusePackagePath | string     |         | Path to the Go package to [reuse]({{< relref "/howtos/code-reuse" >}}) the code from.                                                                                                                      |

## Code

| Attribute              | Type      | Default                         | Description                                                                                                                                               |
|------------------------|-----------|---------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|
| onlyPublish            | bool      | `false`                         | If `true`, generates only the publish code                                                                                                                |
| onlySubscribe          | bool      | `false`                         | If `true`, generates only the subscribe code                                                                                                              |
| disableFormatting      | bool      | `false`                         | If `true`, disables applying the `go fmt` to the generated code                                                                                           |
| targetDir              | string    | `./asyncapi`                    | Target directory name, relative to the current working directory                                                                                          |
| templatesDir           | string    |                                 | Directory with [custom templates]({{< relref "/templating-guide/overview" >}})                                                                            |
| preambleTemplate       | string    | `preamble.tmpl`                 | Preamble template name, used for rendering.                                                                                                               |
| disableImplementations | bool      | `false`                         | If `true`, disables the implementations code generation.                                                                                                  |
| implementationsDir     | string    | `impl/{{ .Manifest.Protocol }}` | [Template expression](#template-expressions) with the implementations directory name, relative to the target directory. Dot is `tmpl.ImplTemplateContext` |

{{% hint info %}}
If both `onlyPublish` and `onlySubscribe` are `false` or omitted, the tool generates both publish and subscribe code.
{{% /hint %}}

## Client

| Attribute        | Type   | Default       | Description                                                               |
|------------------|--------|---------------|---------------------------------------------------------------------------|
| outputFile       | string | `./client`    | Output executable file name                                               |
| outputSourceFile | string | `client.go`   | Name of temporary file with client source code                            |
| keepSource       | bool   | `false`       | If `true`, do not remove the source code file after the build process     |
| goModTemplate    | string | `go.mod.tmpl` | Template name for the client's `go.mod` file                              |
| tempDir          | string |               | Temporary build directory. If empty, use the system's temporary directory |

## Infra

| Attribute  | Type                                | Default                 | Description                                                              |
|------------|-------------------------------------|-------------------------|--------------------------------------------------------------------------|
| serverOpts | [][InfraServerOpt](#infraserveropt) |                         | Additional options for servers generation, such as ServerVariable values |
| engine     | string                              | `docker`                | Target infra engine                                                      |
| outputFile | string                              | `./docker-compose.yaml` | Output file name                                                         |

## InfraServerOpt

| Attribute  | Type                                   | Default | Description                                                                                                                                              |
|------------|----------------------------------------|---------|----------------------------------------------------------------------------------------------------------------------------------------------------------|
| serverName | string                                 |         | **Required**. Server name (key in document)                                                                                                              |
| variables  | map[string]string, []map[string]string |         | [Server variables](https://www.asyncapi.com/docs/concepts/asyncapi-document/variable-url) values. May be a key-value object or list of key-value objects |

## Diagram

| Attribute         | Type                    | Default | Description                                                                                                                    |
|-------------------|-------------------------|---------|--------------------------------------------------------------------------------------------------------------------------------|
| format            | string                  | `svg`   | Output diagram format. Possible values: `svg`, `d2`                                                                            |
| outputFile        | string                  |         | Output file name. Discarded if `--multiple-files` is set                                                                       |
| targetDir         | string                  | `.`     | Directory where to place the output files                                                                                      |
| multipleFiles     | bool                    |         | Generate one diagram file per each source AsyncAPI document instead of all in one diagram. Discards the --output option effect |
| disableFormatting | bool                    |         | Do not run the d2 file formatter. Only applies when output format is d2                                                        |
| channelsCentric   | bool                    |         | Generate the channels-centric diagram                                                                                          |
| serversCentric    | bool                    |         | Generate the servers-centric diagram                                                                                           |
| documentBorders   | bool                    |         | Draw document borders in the diagram                                                                                           |
| d2                | [DiagramD2](#diagramd2) |         | D2 engine options                                                                                                              |

## DiagramD2

| Attribute   | Type                              | Default | Description                                                                   |
|-------------|-----------------------------------|---------|-------------------------------------------------------------------------------|
| engine      | string                            | elk     | D2 layout engine to use. Possible values: `elk`, `dagre`                      |
| direction   | string                            | right   | Diagram draw direction. Possible values: `up`, `down`, `right`, `left`        |
| themeId     | int64                             |         | Theme id to use in diagram. [More info](https://d2lang.com/tour/themes/)      |
| darkThemeId | int64                             |         | Dark theme id to use in diagram. [More info](https://d2lang.com/tour/themes/) |
| pad         | int64                             | 100     | Diagram padding in pixels                                                     |
| sketch      | bool                              |         | Draw diagram in sketchy style                                                 |
| center      | bool                              |         | Center the diagram in the output canvas                                       |                                       |
| scale       | float64                           |         | Scale factor                                                                  |
| elk         | [DiagramD2ELK](#diagramd2elk)     |         | D2 ELK engine options                                                         |
| dagre       | [DiagramD2Dagre](#diagramd2dagre) |         | D2 Dagre engine options                                                       |

## DiagramD2ELK

| Attribute       | Type   | Default                               | Description                                                                                              |
|-----------------|--------|---------------------------------------|----------------------------------------------------------------------------------------------------------|
| algorithm       | string | `layered`                             | Layout algorithm to use. Possible values: `layered`, `force`, `radial`, `mrtree`, `disco`, `rectpacking` |
| nodeSpacing     | int64  | 70                                    | Spacing to be preserved between any pair of nodes of two adjacent layers                                 |
| padding         | string | `[top=50,left=50,bottom=50,right=50]` | Expression of padding to be left to a parent element’s border when placing child elements                |
| edgeSpacing     | int64  | 40                                    | Spacing to be preserved between nodes and edges that are routed next to the node’s layer                 |
| selfLoopSpacing | int64  | 50                                    | Spacing to be preserved between a node and its self loops                                                |

## DiagramD2Dagre

| Attribute | Type  | Default | Description                          |
|-----------|-------|---------|--------------------------------------|
| nodeSep   | int64 | 60      | Number of pixels that separate nodes |
| edgeSep   | int64 | 20      | Number of pixels that separate edges |

# Notes

## Template expressions

Some configuration attributes supports the template expressions, which are evaluated to the string value on tool running.
This allows automatically set a value dynamically, depending on the current object or other context.

The template expression is just a Go [text/template](https://pkg.go.dev/text/template) expression. For example:

```yaml
file: {{.Object.Kind}}s/{{.Object | goIDUpper }}.go
```

Given a particular object, this expression produces the value like `channels/my_channel.go`.

{{% hint info %}}
For more information about the template expressions, see the [templating guide]({{< relref "/templating-guide/overview" >}}).
{{% /hint %}}

## Protocols and implementations

Full list of supported protocols and implementations is available by `list-implementations` CLI subcommand.

## Artifact kind

Every artifact (an intermediate representation object, keeping in memory) generated from the AsyncAPI entity has a kind property. 
Kind is a string enum, for example, "schema", "server", "channel", "operation", etc.

Full list is available in the [artifact.go](https://github.com/bdragon300/go-asyncapi/blob/master/internal/common/artifact.go).
