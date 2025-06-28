---
title: "Partial generation"
weight: 320
description: "How to configure the go-asyncapi to generate only a part of the document entities"
---

# Partial generation

{{% hint info %}}
The features described in this article are useful not only for the code generation, but also for other `go-asyncapi`
outputs, such as generation of client application, infra files, etc.
{{% /hint %}}

## Pub/Sub-only

The easiest way to control the code generation is to use the `--only-pub` and `--only-sub` command-line options 
(or appropriate configuration options) to generate only the publishing or subscribing code, respectively. 
In this case, the entities (e.g. operations), that support the direction opposite to the selected one, will be skipped.

Example:

```shell
go-asyncapi code --only-pub my_asyncapi.yaml
```

or in the configuration file:

```yaml
code:
  onlyPublish: true
```

However, such way might be too broad for your needs. Let's consider the more precise ways.

## x-ignore field

To exclude a particular entity from the code generation, you can use the `x-ignore` special field in the AsyncAPI document.
All references to this entity in the generated Go code are automatically replaced with generic types such as `any`.

The following example shows how to exclude a channel:

```yaml
channels:
  myChannel:
    x-ignore: true
    messages:
      myMessage:
        payload:
          type: string
```

## Code layout

[Code layout]({{< relref "/code-generation/code-layout" >}}) can be used to control the code generation process in a most flexible way.

So, to select the entities, you need to set the rule conditions they should match to. Entities, that didn't match any rule,
will be discarded. If an entity matches several rules, it will be handled by all matching rules separately.

In a rule condition we can check entities name, kind, document location, protocol, etc. The rules can be combined to select
entities by multiple conditions. See [configuration reference]({{< relref "/configuration/reference" >}}) for more details.

Let's consider an example of the code layout configuration:

```yaml
layout:
  - nameRe: "^acme_corp_"
    artifactKinds: ["server", "channel", "operation"]
    reusePackagePath: "github.com/myorg/myproject/pkg/acme_corp"
  - nameRe: "^acme_corp_"
    artifactKinds: ["server", "channel", "operation", "message"]
    except: true
    render:
      file: "{{.Kind}}s/{{.Object | goIDUpper }}.go"
```

In this example, the code for the entities whose names start with **acme_corp_** will be reused from the 
`github.com/myorg/myproject/pkg/acme_corp` package, except for the messages, which will be skipped, because they don't
match any rule condition. Other entities will be handled by the second rule.
