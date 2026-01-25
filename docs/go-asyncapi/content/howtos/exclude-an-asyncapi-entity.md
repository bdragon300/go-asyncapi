---
title: "Exclude an AsyncAPI entity"
weight: 550
description: "How to exclude a particular entity from code generation"
---

# Exclude an AsyncAPI entity

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

See [x-ignore]({{< relref "/asyncapi-specification/special-fields#x-ignore" >}}) description for more details.