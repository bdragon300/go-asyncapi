---
title: "Add a JSONSchema format"
weight: 550
description: "Adding a new JSONSchema format or overriding an existing one"
---

# Add a JSONSchema format or override an existing one

JSON Schema object definition may have a "format" field, that specifies the semantic of the data. 
Along with the "type" field, it defines the data type more precisely. All formats supported by `go-asyncapi` 
out-of-the-box listed in [features page]({{< relref "/features#jsonschema-formats" >}}).

{{% hint note %}}
Unknown JSONSchema formats, that are not built-in nor added as described below, are ignored.
{{% /hint %}}

The following example shows how to add a new custom format `long` for `integer` type.

For that define two following templates and put them to any file, say `my_templates/long_integer.tmpl`:

```gotemplate
{{ define "code/lang/typeFormat/integer/long/definition" -}}
    int64
{{- end }}

{{ define "code/lang/typeFormat/integer/long/usage" -}}
    int64
{{- end }}
```

Then specify the template directory when generating the code:

```bash
go-asyncapi code --template-dir my_templates my_asyncapi.yml
```

Now, the following JSON Schema object will be rendered as `int64` type in Go code:

```yaml
type: integer
format: long
```

{{% hint info %}}
In most cases, it's sufficient to leave the "usage" and "definition" templates identical, but sometimes you may want to 
make them different. For example, the format `foobar` of type `object` here is represented as a struct:

```gotemplate
{{ define "code/lang/typeFormat/object/foobar/definition" -}}
    type {{goID .}} struct {
        Field1 string `json:"field1"`
        Field2 int    `json:"field2"`
    }
{{- end }}

{{ define "code/lang/typeFormat/object/foobar/usage" -}}
    {{ goPkg . }}{{ goID . }}
{{- end }}
```

Learn more about definition and usage code in [templating guide]({{< relref "/templating-guide/overview#definition-and-usage-code" >}}).
{{% /hint %}}
