---
title: "Customize the code layout"
weight: 550
bookToC: true
description: "Customizing the generated code layout"
---

# Customize the code layout

## Abstract code

{{% hint note %}}
For customizing the implementations and utility code layout, see the
next section [Utility and implementations code](#utility-and-implementations-code).
{{% /hint %}}

Many developers may prefer to have the generated code organized in a specific way, so that it is easier to navigate and
maintain. `go-asyncapi` supports customizing the generated code layout, i.e. how the generated code will be broken down by 
packages and files.

Layout is the list of rendering rules that includes condition and options. Every generated artifact is checked 
against every rule, and if it matches, it's rendered according to the options. Artifacts that do not match any rule are 
dropped.

Using the layout rules, you can set up any generated code structure that fits your needs. `go-asyncapi` automatically
handles the imports.

{{% hint warning %}}
Once you have set a custom layout, it's recommended to remove old generated code and regenerate it.
{{% /hint %}}

### Default layout

Default layout is described in [default configuration file](https://github.com/bdragon300/go-asyncapi/blob/master/assets/default_config.yaml).

By default, the code is generated in a way that every entity (like **channels**, **servers**, **schemas**, etc.)
is put into a separate package named after its kind, and a separate file named after the entity name.

{{% tabs "1" %}}
{{% tab "Default configuration" %}}

```yaml
code:
  layout:
    - artifactKinds: [ "schema", "server", "channel", "operation", "message", "parameter" ]
      render:
        file: "{{.Object.Kind}}s/{{.Object | goID }}.go"
    - artifactKinds: [ "security" ]
      render:
        file: "{{.Object.Kind}}/{{.Object | goID }}.go"
```
{{% /tab %}}

{{% tab "Default layout" %}}
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
├── security
│   ├── security_object1.go
│   ├── security_object2.go
│   └── ...
├── ...
└── proto
    ├── kafka
    │   └── ...
    └── ...
```
{{% /tab %}}
{{% /tabs %}}

### Flat package layout

This configuration produces a flat package layout, where all generated code is put into a single package.
Every entity will be placed into a separate file named after its kind and name.

{{% tabs "2" %}}
{{% tab "Configuration" %}}

```yaml
layout:
  - artifactKinds: ["schema", "server", "channel", "operation", "message", "parameter", "security"]
    render:
      file: "{{.Object.Kind}}_{{.Object | goID }}.go"
```

{{% /tab %}}

{{% tab "Layout" %}}

```
target_dir
├── channel_channel1.go
├── server_server1.go
├── schema_schema1.go
├── operation_operation1.go
├── message_message1.go
├── parameter_parameter1.go
├── ...
└── proto
    ├── kafka
    │   └── ...
    └── ...
```

{{% /tab %}}
{{% /tabs %}}

### Splitting into packages by protocol

The next example places all protocol-specific entities into packages named after the protocol (e.g. `kafka`, `amqp`),
while the rest are placed into a separate `common` package.

{{% tabs "3" %}}
{{% tab "Configuration" %}}

```yaml
layout:
  - artifactKinds: ["schema", "parameter", "security"]
    render:
      file: "common/{{.Object | goID }}.go"
  - artifactKinds: ["server", "channel", "operation", "message"]
    protocols: ["kafka"]
    render:
      file: "kafka/{{.Object | goID }}.go"
  - artifactKinds: ["server", "channel", "operation", "message"]
    protocols: ["amqp"]
    render:
      file: "amqp/{{.Object | goID }}.go"
```

{{% /tab %}}

{{% tab "Layout" %}}

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
└── proto
    ├── kafka
    │   └── ...
    └── ...
```

{{% /tab %}}
{{% /tabs %}}

### Splitting into packages by entity name

Here we put the entities (except for schemas) with name matched to regex into a separate package, 
while keep everything else in default layout. [More about regex syntax in Go](https://pkg.go.dev/regexp/syntax)

{{% tabs "4" %}}
{{% tab "Configuration" %}}

```yaml
layout:
  - nameRe: "^acme_corp_"
    artifactKinds: ["server", "channel", "operation", "message", "parameter", "security"]
    render:
      file: "acme/{{.Object | goID }}.go"
  - nameRe: "^acme_corp_"
    artifactKinds: ["schema", "server", "channel", "operation", "message", "parameter"]
    not: true
    render:
      file: "{{.Object.Kind}}s/{{.Object | goID }}.go"
  - nameRe: "^acme_corp_"
    artifactKinds: [ "security" ]
    not: true
    render:
      file: "{{.Object.Kind}}/{{.Object | goID }}.go"
```

{{% /tab %}}

{{% tab "Layout" %}}

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
└── proto
    ├── kafka
    │   └── ...
    └── ...
```

{{% /tab %}}
{{% /tabs %}}

### Only schemas

The following example shows how to generate only the schemas, ignoring all other entities.

{{% tabs "5" %}}
{{% tab "Configuration" %}}

```yaml
layout:
  - artifactKinds: ["schema"]
    render:
      file: "schemas/{{.Object | goID }}.go"
```

{{% /tab %}}

{{% tab "Layout" %}}

```
target_dir
├── schemas
│   ├── Schema1.go
│   ├── Schema2.go
│   └── Schema3.go
└── proto
    ├── kafka
    │   └── ...
    └── ...
```

{{% /tab %}}
{{% /tabs %}}

### Everything in one file

{{% tabs "6" %}}
{{% tab "Configuration" %}}

```yaml
layout:
  - artifactKinds: ["schema", "server", "channel", "operation", "message", "parameter", "security"]
    render:
      file: all.go
```

{{% /tab %}}

{{% tab "Layout" %}}

```
target_dir
├── all.go
└── proto
    ├── kafka
    │   └── ...
    └── ...
```

{{% /tab %}}
{{% /tabs %}}

## Utility and implementations code

Path to the implementations code (3rd-party libraries code) and utility code (interfaces, util types, etc.) are 
configurable separately from the layout because they are not generated from AsyncAPI entities, so they do not match any layout rules.
Implementations are appeared only if `go-asyncapi` has built-in implementation for this protocol.

### Default layout

{{% tabs "7" %}}
{{% tab "Default configuration" %}}

```yaml
code:
  util:
    directory: "proto/{{ .Protocol }}"
  implementation:
    directory: "proto/{{ .Protocol }}"
```
{{% /tab %}}

{{% tab "Default layout" %}}

```
target_dir
└── proto
    ├── kafka
    │   ├── util.go
    │   └── implementation.go
    ├── amqp
    │   ├── util.go
    │   └── implementation.go
    └── ...
```
{{% /tab %}}
{{% /tabs %}}

### Implementations and utility code in separate packages

{{% tabs "8" %}}
{{% tab "Configuration" %}}

```yaml
code:
  util:
    directory: "impl/{{ .Protocol }}"
  implementation:
    directory: "util/{{ .Protocol }}"
```
{{% /tab %}}

{{% tab "Layout" %}}

```
target_dir
├── impl
│   ├── kafka
│   │   └── ...
│   ├── amqp
│   │   └── ...
│   └── ...
└── util
    ├── kafka
    │   └── ...
    ├── amqp
    │   └── ...
    └── ...
```
{{% /tab %}}
{{% /tabs %}}

### Implementations package named after library

For example, we can put implementations into subpackages named after the implementation name instead of protocol name.

{{% tabs "9" %}}
{{% tab "Configuration" %}}

```yaml
code:
  util:
    directory: "impl/{{ .Manifest.Name }}"
  implementation:
    directory: "util/{{ .Protocol }}"
```
{{% /tab %}}

{{% tab "Layout" %}}

```
target_dir
├── impl
│   └── github.com
│       ├── twmb
│       │   └── franz-go
│       │       └── ...
│       └── rabbitmq
│           └── amqp091-go
│               └── ...
└── util
    ├── kafka
    │   └── ...
    ├── amqp
    │   └── ...
    └── ...
```
{{% /tab %}}
{{% /tabs %}}

### Disable all protocols

To disable the generation of implementations code, you can use `--disable-implementations` CLI flag or enable the 
setting in the configuration file:

{{% hint note %}}
Utility code cannot be disabled because the abstract code depends on it.
{{% /hint %}}

```yaml
code:
  implementation:
    disable: true
```

### Disable only a particular protocol

It's possible to avoid generating the implementations code only for a particular protocol:

```yaml
code:
  implementation:
    custom:
      - protocol: kafka
        disable: true
```