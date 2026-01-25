---
title: "Configuration reference"
weight: 800
description: "go-asyncapi configuration reference"
---

# Configuration reference

Configuration is set in a single YAML file, that may contain the attributes described below. 
All attributes are optional, and if not set, the tool uses the default values.

{{% hint tip %}}
By default, `go-asyncapi` reads the configuration file named `go-asyncapi.yaml` or `go-asyncapi.yml` in the 
current working directory if it exists. You may set you own location by passing the `-c` CLI option.
{{% /hint %}}

{{% hint tip %}}
See also [default configuration file](https://github.com/bdragon300/go-asyncapi/blob/master/assets/default_config.yaml)
{{% /hint %}}

## Config

| Attribute       | Type                                | Default                                                                             | Description                                                                                              |
|-----------------|-------------------------------------|-------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------|
| configVersion   | int                                 | `1`                                                                                 | Configuration version. Now only the `1` is valid value                                                   |
| projectModule   | string                              |                                                                                     | Project module name for the generated code. If empty, takes from go.mod in the current working directory |
| runtimeModule   | string                              | `github.com/bdragon300/go-asyncapi/run`                                             | Path to runtime module with auxiliary code                                                               |
| templatesDir    | string                              |                                                                                     | Directory with [custom templates]({{< relref "/templating-guide/overview" >}})                           |
| locator         | [Locator](#locator)                 |                                                                                     | [Reference locator]({{< relref "/asyncapi-specification/references#reference-locator" >}}) settings      |
| code            | [Code](#code)                       |                                                                                     | Code generation settings                                                                                 |
| client          | [Client](#client)                   |                                                                                     | No-code client building settings                                                                         |
| infra           | [Infra](#infra)                     |                                                                                     | Infra files generation settings                                                                          |
| diagram         | [Diagram](#diagram)                 |                                                                                     | Diagram generation settings                                                                              |
| ui              | [UI](#ui)                           |                                                                                     | Web UI generation settings                                                                               |

## Locator

| Attribute             | Type          | Default | Description                                                                                                                                           |
|-----------------------|---------------|---------|-------------------------------------------------------------------------------------------------------------------------------------------------------|
| allowRemoteReferences | bool          | `false` | `true` allows reading remote `$ref`s (denied by default by security reasons)                                                                          |
| rootDirectory         | string        |         | The base directory to locate the files. Not used if empty.                                                                                            |
| timeout               | time.Duration | `30s`   | Reference resolving timeout (sets both the HTTP request timeout and command execution timeout)                                                        |
| command               | string        |         | Shell command of [custom locator]({{< relref "/asyncapi-specification/references#custom-reference-locator" >}}). If empty, uses the built-in locator. |

{{% hint tip %}}
For more information about the reference locator, see the [References]({{< relref "/asyncapi-specification/references" >}}) page.
{{% /hint %}}

## Code

| Attribute              | Type                                | Default                                                                             | Description                                                                                                                                               |
|------------------------|-------------------------------------|-------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------------------------------------------|
| onlyPublish            | bool                                | `false`                                                                             | If `true`, generates only the publish code                                                                                                                |
| onlySubscribe          | bool                                | `false`                                                                             | If `true`, generates only the subscribe code                                                                                                              |
| disableFormatting      | bool                                | `false`                                                                             | If `true`, disables applying the `go fmt` to the generated code                                                                                           |
| targetDir              | string                              | `./asyncapi`                                                                        | Target directory name, relative to the current working directory                                                                                          |
| layout                 | [][Layout](#layout)                 | [Default layout]({{< relref "/howtos/customize-the-code-layout#default-layout" >}}) | Generated code layout rules                                                                                                                               |
| preambleTemplate       | string                              | `preamble.tmpl`                                                                     | Preamble template name, used for rendering.                                                                                                               |
| util                   | [Util](#util)                       |                                                                                     | Utility code generation settings                                                                                                                          |
| implementation         | [][Implementation](#implementation) |                                                                                     | Per-protocol implementations generation settings                                                                                                          |

{{% hint info %}}
If both `onlyPublish` and `onlySubscribe` are `false`, the tool generates both publish and subscribe code.
{{% /hint %}}

## Layout

{{% hint tip %}}
For more info see the [Code layout]({{< relref "/howtos/customize-the-code-layout" >}}) page.
{{% /hint %}}

| Attribute        | Type                             | Default | Description                                                                                 |
|------------------|----------------------------------|---------|---------------------------------------------------------------------------------------------|
| nameRe           | string                           |         | Condition: regex to match the entity name                                                   |
| artifactKinds    | []string                         |         | Condition: match to one of [artifact kinds](#artifact-kind) in list                         |
| moduleURLRe      | string                           |         | Condition: regex to match the document URL or path                                          |
| pathRe           | string                           |         | Condition: regex to match the artifact path inside a document (e.g. `#/channels/myChannel`) |
| protocols        | []string                         |         | Condition: match to any [protocol](#protocols) in list                                      |
| not              | bool                             | `false` | If `true`, apply NOT operations to the match                                                |
| render           | [LayoutRender](#layoutrender)    |         | Layout rule rendering options                                                               |

{{% hint info %}}
Several conditions in one layout rule are joined via **AND** operation.
{{% /hint %}}

## LayoutRender

| Attribute        | Type     | Default     | Description                                                                                                         |
|------------------|----------|-------------|---------------------------------------------------------------------------------------------------------------------|
| file             | string   |             | **Required**. [Template expression](#template-expressions) with output file name. Dot is `tmpl.CodeTemplateContext` |
| package          | string   |             | Package name. If empty, takes from file's directory name                                                            |
| protocols        | []string |             | [Protocols](#protocols) that are allowed to render. Empty list means no restriction                                 |
| template         | string   | `main.tmpl` | Root template name                                                                                                  |

## Util

| Attribute | Type                            | Default                  | Description                                                                                                                                       |
|-----------|---------------------------------|--------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------|
| directory | string                          | `proto/{{ .Protocol }}`  | [Template expression](#template-expressions) with output directory name, relative to the target directory. Dot is `tmpl.CodeExtraTemplateContext` |
| custom    | [][UtilProtocol](#utilprotocol) |                          | Per-protocol util code customization options                                                                                                      |

## UtilProtocol

| Attribute          | Type       | Default | Description                          |
|--------------------|------------|---------|--------------------------------------|
| protocol           | string     |         | **Required**. [Protocol](#protocols) |
| templateDirectory  | string     |         | Directory with custom templates      |

## Implementation

| Attribute | Type                                                 | Default                  | Description                                                                                                                                                         |
|-----------|------------------------------------------------------|--------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| directory | string                                               | `proto/{{ .Protocol }}`  | [Template expression](#template-expressions) with output directory name for all protocols, relative to the target directory. Dot is `tmpl.CodeExtraTemplateContext` |
| disable   | bool                                                 | `false`                  | If `true`, disable generating the implementations for all protocols                                                                                                 |
| custom    | [][ImplementationProtocol](#implementationprotocol)  |                          | Per-protocol implementation code customization options                                                                                                              |

## ImplementationProtocol

| Attribute         | Type       | Default | Description                                                                                           |
|-------------------|------------|---------|-------------------------------------------------------------------------------------------------------|
| protocol          | string     |         | **Required**. [Protocol](#protocols)                                                                  |
| name              | string     |         | Implementation name.                                                                                  |
| disable           | bool       | `false` | If `true`, disables the implementation code generation. Overrides the global implementation `disable` |
| templateDirectory | string     |         | Directory with custom templates                                                                       |
| package           | string     |         | Package name if it differs from the directory name                                                    |

## Client

| Attribute        | Type   | Default       | Description                                                               |
|------------------|--------|---------------|---------------------------------------------------------------------------|
| outputFile       | string | `./client`    | Output executable file name                                               |
| outputSourceFile | string | `client.go`   | Name of temporary file with client source code                            |
| keepSource       | bool   | `false`       | If `true`, do not remove the source code file after the build finished    |
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
| outputFile        | string                  |         | Output file name. If empty, same as name of an input document. Has no effect on rendering multiple files.                      |
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
| engine      | string                            | `elk`   | D2 layout engine to use. Possible values: `elk`, `dagre`                      |
| direction   | string                            | `right` | Diagram draw direction. Possible values: `up`, `down`, `right`, `left`        |
| themeId     | int64                             |         | Theme id to use in diagram. [More info](https://d2lang.com/tour/themes/)      |
| darkThemeId | int64                             |         | Dark theme id to use in diagram. [More info](https://d2lang.com/tour/themes/) |
| pad         | int64                             | `100`   | Diagram padding in pixels                                                     |
| sketch      | bool                              |         | Draw diagram in sketchy style                                                 |
| center      | bool                              |         | Center the diagram in the output canvas                                       |                                       |
| scale       | float64                           |         | Scale factor                                                                  |
| elk         | [DiagramD2ELK](#diagramd2elk)     |         | D2 ELK engine options                                                         |
| dagre       | [DiagramD2Dagre](#diagramd2dagre) |         | D2 Dagre engine options                                                       |

## DiagramD2ELK

| Attribute       | Type   | Default                               | Description                                                                                              |
|-----------------|--------|---------------------------------------|----------------------------------------------------------------------------------------------------------|
| algorithm       | string | `layered`                             | Layout algorithm to use. Possible values: `layered`, `force`, `radial`, `mrtree`, `disco`, `rectpacking` |
| nodeSpacing     | int64  | `70`                                  | Spacing to be preserved between any pair of nodes of two adjacent layers                                 |
| padding         | string | `[top=50,left=50,bottom=50,right=50]` | Expression of padding to be left to a parent element’s border when placing child elements                |
| edgeSpacing     | int64  | `40`                                  | Spacing to be preserved between nodes and edges that are routed next to the node’s layer                 |
| selfLoopSpacing | int64  | `50`                                  | Spacing to be preserved between a node and its self loops                                                |

## DiagramD2Dagre

| Attribute | Type  | Default | Description                          |
|-----------|-------|---------|--------------------------------------|
| nodeSep   | int64 | `60`    | Number of pixels that separate nodes |
| edgeSep   | int64 | `20`    | Number of pixels that separate edges |

## UI

| Attribute     | Type   | Default | Description                                                                                                                                              |
|---------------|--------|---------|----------------------------------------------------------------------------------------------------------------------------------------------------------|
| outputFile    | string |         | Output HTML file name. If empty, same input document name with ".html" extension. Has no effect on "listening" mode                                      |
| listen        | bool   | `false` | Enable listening mode using the internal web server                                                                                                      |
| listenAddress | string | `:8090` | Host/port to serve the UI in listening mode                                                                                                              |
| listenPath    | string | `/`     | URL base path to serve the UI in listening mode                                                                                                          |
| bundle        | bool   | `false` | Make a bundle. See also [Bundling the assets]({{< relref "/commands/ui#bundling-the-assets" >}})                                                         |
| bundleDir     | string |         | If not empty, get assets to bundle from directory instead of using default assets. See also [Custom assets]({{< relref "/commands/ui#custom-assets" >}}) |


# Notes

## Template expressions

Some configuration attributes support the template expressions, which are evaluated to the string value on tool running.
This enables to set a value dynamically depending on the current object or context.

The template expression is just a Go [text/template](https://pkg.go.dev/text/template) expression. For example:

```yaml
file: {{.Object.Kind}}s/{{ goID .Object }}.go
```

Given a particular object, this expression produces the value like `channels/MyChannel.go`.

{{% hint info %}}
For more information about the template expressions, see the [templating guide]({{< relref "/templating-guide/overview" >}}).
{{% /hint %}}

## Protocols and implementations

Full list of implementations built-in for particular protocols is available by `list-implementations` CLI subcommand.

## Artifact kind

{{% hint tip %}}
For more info about artifacts see [Internals]({{< relref "/internals" >}}) article.
{{% /hint %}}

Every artifact generated from the AsyncAPI entity has a kind, which is a string enum, for example, "schema", "server", "channel", "operation", etc.

Full list is available in the [artifact.go](https://github.com/bdragon300/go-asyncapi/blob/master/internal/common/artifact.go).
