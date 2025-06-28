---
title: "Code reuse"
weight: 340
description: "Reusing the existing Go code in the generated code"
---

# Code reuse

While generating the code, you may want to reuse the code you already have. There are two ways to do that.

## Reuse a schema type

Jsonschema document entities may generate the struct, array, built-in Go types. 
But you can mark any jsonschema to reuse your existing Go type instead, which may be a local type or type imported 
from an external module.
All usages of the original type in the generated code will be replaced automatically with the specified type.

See the [x-go-type extra field]({{< relref "/asyncapi-specification/special-fields#x-go-type" >}}) for details.

{{% details "Example" open %}}
```yaml
components:
  schemas:
    MySchema:
      x-go-type: MyOwnType
      type: object
      properties:
        myField:
          type: string
```
{{% /details %}}

## Reuse the code blocks

It's possible to mark the reuse the [code layout]({{< relref "/code-generation/code-layout" >}}) rule to reuse the code
from the existing Go package instead of generating the code. 

All usages of names from the reused code don't change in other places in the generated code, except that they
are be imported from the reused package.

The following rule shows how to reuse the code of all parameters from `github.com/myorg/mymodule/schemas` package:

```yaml
layout:
  - artifactKinds: ["parameter"]
    reusePackagePath: "github.com/myorg/mymodule/schemas"
#...
```

All available fields are described in [configuration reference]({{< relref "/configuration/reference" >}}).
