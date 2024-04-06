---
title: "Model"
weight: 440
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
[Features]({{< relref "/docs/features" >}}) page.

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


## x-nullable

Extra field `x-nullable` forcibly marks a model/field as nullable. By default, the field is nullable if it can be
`null` or needed to be checked for `nil` in the generated code, i.e. marked as required. Nullable fields are generated
as pointers in Go code.

{{< details "Example" >}}
{{< tabs "2" >}}
{{< tab "Definition" >}}
```yaml
components:
  schemas:
    myModel:
      type: object
      required:
        - id
      properties:
        id:
          type: integer
        name:
          type: string
          x-nullable: true
        description:
          type: string
```
{{< /tab >}}

{{< tab "Produced code" >}}
```go
package models

type MyModel struct {
    ID          *int     `json:"id"`
    Name        *string  `json:"name"`
    Description string   `json:"description"`
}
```
{{< /tab >}}
{{< /tabs >}}
{{< /details >}}

## x-go-name

This extra field is used to explicitly set the name of the model in generated code. By default, for named models the 
Go name is generated from schema's name, for anonymous schema the Go name is deduced from the parent objects 
where the object is defined.

{{< details "Example" >}}
{{< tabs "3" >}}
{{< tab "Definition" >}}
```yaml
components:
  schemas:
    myModel:
      x-go-name: FooBar
      type: object
      properties:
        id:
          x-go-name: FooID
          type: integer
        name:
          type: string
```
{{< /tab >}}

{{< tab "Produced code" >}}
```go

//...

type FooBar struct {
    FooID   int `json:"id"`
    Name string `json:"name"`
}

//...
```
{{< /tab >}}
{{< /tabs >}}
{{< /details >}}

## x-go-type

This extra field overrides a model type. This is useful when you want to use a custom type instead of generated one.
If the `x-go-type` is set, the tool ignores the object definition and will import and use the specified type
instead. It can be any built-in type, a type from the standard library, third-party libraries, and
other generated types.

`x-go-type` can have string value with type name, e.g. `x-go-type: string`. In this case, the tool just 
substitutes the type name everywhere where the model is used. Suitable for built-in types and types from the 
same package as the model.

The advanced way is to use `x-go-type` as object. Possible subfields are:

* `name` -- the type name. Required.
* `import` -- the import path for the type. Optional. Can be either a package name in generated code, e.g. `"messages"`
  or a full import path, e.g. `"github.com/myorg/mylib"`. If empty, the tool just uses the type name.
* `embedded` -- if **true**, the tool will embed a type in the generated model instead of replacing it. Optional. 
  Default is **false**.
* `hint` -- various hints how this type should be used. Optional. Possible values are:
  * `pointer` -- if **true**, then objects of this type should be used as pointers in the generated code. If **false**
    (default) being a pointer depends on context.
  * `kind` -- the kind of the type. Now it's supported only the `"interface"`, which denotes that the type is Go 
    interface. This means, for example, that we can't get the address of the object of this type.

{{< details "Example" >}}
```yaml
components:
  schemas:
    myModel:
      x-go-type:
        name: MyCustomType
        import: "github.com/myorg/mylib"
        embedded: true  # `github.com/myorg/mylib.MyCustomType` will be embedded to `MyModel`
        hint:
          kind: "interface"  # Treat the `MyCustomType` as interface
      type: object
      properties:
        id:
          type: integer
          x-go-type: int8  # `ID` field will have int8 type
        name:
          type: string
        server:
          type: object
            x-go-type:
              name: MyServer  # `MyServer` from the `servers` generated package will be used
              import: "servers"
```
{{< /details >}}

## x-go-tags and x-go-tags-values

You can add your own tags to the generated model fields using the `x-go-tags` and `x-go-tags-values` extra fields.
By default, the tags are generated based on the content type of the message where the model is used.

To add tags that is meant to have the same values as other ones, pass a list to the `x-go-tags` field.

{{< details "Example" >}}
{{< tabs "4" >}}
{{< tab "Definition" >}}
```yaml
components:
  schemas:
    myModel:
      type: object
      properties:
        id:
          x-go-tags:
            - spam
            - eggs
          type: integer
        name:
          type: string
```
{{< /tab >}}

{{< tab "Produced code" >}}
```go
package models

type MyModel struct {
    ID   int    `json:"id" spam:"id" eggs:"id"`
    Name string `json:"name"`
}
```
{{< /tab >}}
{{< /tabs >}}
{{< /details >}}

To add certain tags with certain values, pass a map of values to the `x-go-tags` field. If such tag is already present
in the generated code, the value will be replaced.

{{< details "Example" >}}
{{< tabs "5" >}}
{{< tab "Definition" >}}
```yaml
components:
  schemas:
    myModel:
      type: object
      properties:
        id:
          x-go-tags:
            spam: "foo"
            eggs: "bar"
            json: "baz"
          type: integer
        name:
          type: string
```
{{< /tab >}}

{{< tab "Produced code" >}}
```go
package models

type MyModel struct {
    ID   int    `json:"baz" spam:"foo" eggs:"bar"`
    Name string `json:"name"`
}
```
{{< /tab >}}
{{< /tabs >}}
{{< /details >}}

You also add flags to all tags, using the `x-go-tags-values` field.

{{< details "Example" >}}
{{< tabs "6" >}}
{{< tab "Definition" >}}
```yaml
components:
  schemas:
    myModel:
      type: object
      properties:
        id:
          x-go-tags-values:
            - omitempty
          x-go-tags:
            - spam
            - eggs
          type: integer
        name:
          type: string
```
{{< /tab >}}

{{< tab "Produced code" >}}
```go
package models

type MyModel struct {
    ID   int    `json:"id,omitempty" spam:"id,omitempty" eggs:"id,omitempty"`
    Name string `json:"name"`
}
```
{{< /tab >}}
{{< /tabs >}}
{{< /details >}}

## x-ignore

If this extra field it set to **true**, the model will not be generated. All references
to this model in the generated code (if any) are replaced by Go `any` type.

{{< details "Example" >}}
```yaml
components:
  schemas:
    myModel:
      x-ignore: true
      type: object
      properties:
        id:
          type: integer
        name:
          type: string
```
{{< /details >}}