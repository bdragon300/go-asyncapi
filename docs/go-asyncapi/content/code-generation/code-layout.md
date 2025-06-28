---
title: "Code layout"
weight: 330
description: "Customizing the generated code layout"
---

# Code layout

Many developers may prefer to have the generated code organized in a specific way, so that it is easier to navigate and
maintain. `go-asyncapi` supports customizing the generated code layout, i.e. how the generated code will be broken down by 
packages and files. Others may want to customize the generation process only for a particular AsyncAPI objects, 
keeping everything as-is.

Layout is a series of rules that includes condition and rule fields. During the generation process, every entity is 
checked against every rule, and if it matches, it's rendered according to the rule.

Entities that do not match any rule are ignored. This means, for example, if you would have empty rules list, the tool 
would not generate any code at all.

So, the layout rules may be used for the following purposes:

* Grouping the code by packages and files in any way. `go-asyncapi` automatically recalculates the
  correct code imports in the generated code once layout has changed.
* Reusing the code from the existing modules. Technically, the entities Go definitions matched by this rule are 
  not generated, but instead the tool will add a Go import from the specified module.
* Setting the custom code template specific for a rule. Useful when a particular entity should be rendered in a specific way.

{{% hint warning %}}
Once you have set a custom layout, it's recommended to remove old generated code and regenerate it.
{{% /hint %}}

{{% hint info %}}
For more information see the [Configuration]({{< relref "/configuration" >}}) page.
{{% /hint %}}

## Default layout

By default, the code is generated in a way that every entity type (like **channels**, **servers**, **schemas**, etc.)
is put into a separate package, and every single entity is put into a separate file.

{{< tabs "1" >}}
{{< tab "Default configuration" >}}

```yaml
layout:
  - artifactKinds: ["schema", "server", "channel", "operation", "message", "parameter"]
    render:
      file: "{{.Object.Kind}}s/{{.Object | goIDUpper }}.go"
```
{{< /tab >}}

{{< tab "Default layout" >}}
```
target_dir
├── channels
│   ├── channel1.go
│   ├── channel2.go
│   └── ...
├── servers
│   ├── server1.go
│   ├── server2.go
│   └── ...
├── schemas
│   ├── schema1.go
│   ├── schema2.go
│   └── ...
├── ...
└── impl
    ├── kafka
    │   └── ...
    └── ...
```
{{< /tab >}}
{{< /tabs >}}

## Examples

The following example produces the layout, where everything is put into target directory in one package as separate 
files with name `<entity_type>_<entity_name>.go`, e.g. `channel_channel1.go`, `server_server1.go`, etc.

{{< tabs "2" >}}
{{< tab "Configuration" >}}

```yaml
layout:
  - artifactKinds: ["schema", "server", "channel", "operation", "message", "parameter"]
    render:
      file: "{{.Kind}}_{{.Object | goIDUpper }}.go"
```

{{< /tab >}}

{{< tab "Layout" >}}

```
target_dir
├── channel_channel1.go
├── server_server1.go
├── schema_schema1.go
├── operation_operation1.go
├── message_message1.go
├── parameter_parameter1.go
├── ...
└── impl
    ├── kafka
    │   └── ...
    └── ...
```

{{< /tab >}}
{{< /tabs >}}

The next example places all protocol-agnostic entities into the `common` package, and protocol-specific entities into 
packages with the protocol name.

{{< tabs "3" >}}
{{< tab "Configuration" >}}

```yaml
layout:
  - artifactKinds: ["schema", "parameter"]
    render:
      file: "common/{{.Object | goIDUpper }}.go"
  - artifactKinds: ["server", "channel", "operation", "message"]
    protocols: ["kafka"]
    render:
      file: "kafka/{{.Object | goIDUpper }}.go"
  - artifactKinds: ["server", "channel", "operation", "message"]
    protocols: ["amqp"]
    render:
      file: "amqp/{{.Object | goIDUpper }}.go"
```

{{< /tab >}}

{{< tab "Layout" >}}

```
target_dir
├── common
│   ├── Schema1.go
│   ├── Parameter1.go
├── kafka
│   ├── Server1.go
│   ├── Channel1.go
│   ├── Operation1.go
│   └── Message1.go
├── amqp
│   ├── Server2.go
│   ├── Channel2.go
│   ├── Operation2.go
│   └── Message2.go
├── ...
└── impl
    ├── kafka
    │   └── ...
    └── ...
```

{{< /tab >}}
{{< /tabs >}}

What about to put entities (except for schemas) only with specific name prefix into a separate package, 
while keep everything else in default layout? [More about regex syntax in Go](https://pkg.go.dev/regexp/syntax)

{{< tabs "4" >}}
{{< tab "Configuration" >}}

```yaml
layout:
  - nameRe: "^acme_corp_"
    artifactKinds: ["server", "channel", "operation", "message", "parameter"]
    render:
      file: "acme/{{.Object | goIDUpper }}.go"
  - nameRe: "^acme_corp_"
    artifactKinds: ["schema", "server", "channel", "operation", "message", "parameter"]
    except: true
    render:
      file: "{{.Kind}}s/{{.Object | goIDUpper }}.go"
```

{{< /tab >}}

{{< tab "Layout" >}}

```
target_dir
├── acme
│   ├── Server1.go
│   ├── Channel1.go
│   ├── Operation1.go
│   └── Message1.go
├── schemas
│   ├── Schema2.go
│   ├── Schema3.go
│   └── ...
├── servers
│   ├── Server2.go
│   ├── Server3.go
│   └── ...
├── channels
│   ├── Channel2.go
│   ├── Channel3.go
│   └── ...
├── operations
│   ├── Operation2.go
│   ├── Operation3.go
│   └── ...
├── messages
│   ├── Message2.go
│   ├── Message3.go
│   └── ...
└── impl
    ├── kafka
    │   └── ...
    └── ...
```

{{< /tab >}}
{{< /tabs >}}

The following example shows how to generate only the schemas, ignoring all other entities.

{{< tabs "5" >}}
{{< tab "Configuration" >}}

```yaml
layout:
  - artifactKinds: ["schema"]
    render:
      file: "schemas/{{.Object | goIDUpper }}.go"
```

{{< /tab >}}

{{< tab "Layout" >}}

```
target_dir
├── schemas
│   ├── Schema1.go
│   ├── Schema2.go
│   └── Schema3.go
└── impl
    ├── kafka
    │   └── ...
    └── ...
```

{{< /tab >}}
{{< /tabs >}}

Finally, let's put everything into one file.

{{< tabs "6" >}}
{{< tab "Configuration" >}}

```yaml
layout:
  - artifactKinds: ["schema", "server", "channel", "operation", "message", "parameter"]
    render:
      file: all.go
```

{{< /tab >}}

{{< tab "Layout" >}}

```
target_dir
├── all.go
└── impl
    ├── kafka
    │   └── ...
    └── ...
```

{{< /tab >}}
{{< /tabs >}}

## Implementations

Implementations code layout is configured separately, because it is not related to the AsyncAPI entities.
By default, the code is put to `impl` package, and each protocol has its own subpackage.

{{< tabs "7" >}}
{{< tab "Default configuration" >}}

```yaml
code:
  implementationsDir: "impl/{{ .Manifest.Protocol }}"
```
{{< /tab >}}

{{< tab "Default layout" >}}

```
target_dir
└── impl
    ├── kafka
    │   └── ...
    ├── amqp
    │   └── ...
    └── ...
```
{{< /tab >}}
{{< /tabs >}}

For example, we can change the layout to put implementations into subpackages named after the implementation name instead
of protocol name.

{{< tabs "8" >}}
{{< tab "Configuration" >}}

```yaml
code:
  implementationsDir: "impl/{{ .Manifest.Name }}"
```
{{< /tab >}}

{{< tab "Layout" >}}

```
target_dir
└── impl
    ├── franz-go
    │   └── ...
    ├── amqp091-go
    │   └── ...
    └── ...
```
{{< /tab >}}
{{< /tabs >}}

To disable the injecting the implementations code, you can use `--disable-implementations` CLI flag or enable the 
setting in the configuration file:

```yaml
code:
  disableImplementations: true
```
