---
title: "Model"
weight: 1
# bookFlatSection: false
# bookToc: true
# bookHidden: false
# bookCollapseSection: false
# bookComments: false
# bookSearchExclude: false
---

# Model

## Overview

Models are just structs generated from [JSONSchema](https://json-schema.org/) definitions in the `components.schemas` 
AsyncAPI section. 

The tool also considers all content types of the messages where a particular model is used. So, for example, if a model
is met in messages of JSON and YAML content types, the field tags will be: `json:"fieldName" yaml:"fieldName"`.

The JSONSchema features and content types supported by the tool are described on the 
[Features]({{< relref "/docs/overview/features" >}}) page.

The generated code is placed in the `models` package by default.

{{< details "Minimal example" >}}
{{< tabs "1" >}}
{{< tab "Definition" >}}
```yaml
components:
  messages:
    myMessage:
      payload:
        $ref: '#/components/schemas/myModel'

    myMessage2:
      payload:
        $ref: '#/components/schemas/myModel'
      contentType: application/yaml

  schemas:
    myModel:
      type: object
      properties:
        id:
          type: integer
        name:
          type: string
```
{{< /tab >}}

{{< tab "Produced code" >}}
```go
package models

type MyModel struct {
	ID   int    `json:"id" yaml:"id"`
	Name string `json:"name" yaml:"name"`
}
```
{{< /tab >}}
{{< /tabs >}}
{{< /details >}}