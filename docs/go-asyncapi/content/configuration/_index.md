---
title: "Configuration reference"
weight: 800
bookToC: true
description: "go-asyncapi configuration reference"
---

# Configuration reference

Configuration is set in a single YAML file, that may contain the attributes described below. 
All attributes are optional, and if not set, the tool uses the default values.

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

## Layout

| Attribute        | Type                               | Default | Description                                                                                 |
|------------------|------------------------------------|---------|---------------------------------------------------------------------------------------------|
| nameRe           | string                             |         | Condition: regex to match the entity name                                                   |
| artifactKinds    | []string                           |         | Condition: match to one of [artifact kinds](#artifact-kind) in list                         |
| moduleURLRe      | string                             |         | Condition: regex to match the document URL or path                                          |
| pathRe           | string                             |         | Condition: regex to match the artifact path inside a document (e.g. `#/channels/myChannel`) |
| protocols        | []string                           |         | Condition: match to any [protocol](#protocols-and-implementations) in list                  |
| render           | [LayoutRender](#layoutrender)      |         | Render settings                                                                             |
| reusePackagePath | string                             |         | Path to the Go package to [reuse]({{< relref "/howtos/code-reuse" >}}) the code from.       |

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

| Attribute               | Type          | Default | Description                                                                                                                                            |
|-------------------------|---------------|---------|--------------------------------------------------------------------------------------------------------------------------------------------------------|
| allowRemoteReferences   | bool          | `false` | `true` allows reading remote `$ref`s (denied by default by security reasons)                                                                           |
| searchDirectory         | string        | `.`     | Directory to search for local files referenced by `$ref`s.                                                                                             |
| timeout                 | time.Duration | `30s`   | Reference resolving timeout (sets both the HTTP request timeout and command execution timeout)                                                         |
| command                 | string        |         | Shell command of [custom locator]({{< relref "/asyncapi-specification/references#custom-reference-locator" >}}). If empty, uses the built-in locator.  |

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

| Attribute        | Type                          | Default                 | Description                                                                    |
|------------------|-------------------------------|-------------------------|--------------------------------------------------------------------------------|
| servers          | [][InfraServer](#infraserver) |                         | Additional arguments for AsyncAPI server entities, used to generate the result |
| format           | string                        | `docker`                | Output file format                                                             |
| outputFile       | string                        | `./docker-compose.yaml` | Output file name                                                               |

## InfraServer

| Attribute        | Type                                   | Default | Description                                                                                                                                              |
|------------------|----------------------------------------|---------|----------------------------------------------------------------------------------------------------------------------------------------------------------|
| name             | string                                 |         | **Required**. Server name                                                                                                                                |
| variables        | map[string]string, []map[string]string |         | [Server variables](https://www.asyncapi.com/docs/concepts/asyncapi-document/variable-url) values. May be a key-value object or list of key-value objects |

# Notes

## Template expressions

Some configuration attributes supports the template expressions, which are evaluated to the string value on tool running.
This allows automatically set a value dynamically, depending on the current object or other context.

The template expression is just a Go [text/template](https://pkg.go.dev/text/template) expression. Couple of examples:

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
