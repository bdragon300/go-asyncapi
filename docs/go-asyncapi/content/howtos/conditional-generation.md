---
title: "Conditional code generation"
weight: 320
description: "How to conditionally generate code for AsyncAPI entities"
---

# Conditional code generation

{{% hint info %}}
The features described in this article are useful not only for the code generation, but also for other `go-asyncapi`
outputs, such as generation of client application, infra files, etc.
{{% /hint %}}

## Pub/Sub-only generation

The easiest way to control the code generation is to use the `--only-pub` and `--only-sub` command-line options 
(or appropriate configuration options) to generate only the publishing or subscribing code, respectively. 
In this case, the entities (e.g. operations), that support the direction opposite to the selected one, will be skipped.

Example:

```shell
go-asyncapi code --only-pub my_asyncapi.yaml
```

However, such way to control the codegen process might be too broad for your needs. Let's consider the more precise ways.

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

You can control what entities should produce the code by configuring the 
[code layout]({{< relref "/code-generation/code-layout" >}}).

Code layout is a list of rules, each of which describes how to render the code of entities' artifacts, matched to a rule's
conditions.

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
