---
title: "JSONSchema"
weight: 760
description: "Using JSONSchema object in AsyncAPI specification"
---

# Format

JSON Schema object may contain a "format" field, that specifies the semantic of the data. Along with the "type" field,
it defines the data type more precisely.

Example:

```yaml
type: string
format: date-time
```

`go-asyncapi` supports several formats out of the box, see [features page]({{< relref "/features#jsonschema-formats" >}}).

## Adding a new format

{{% hint tip %}}
You can override any existing format the same way.
{{% /hint %}}

You can add an implementation for any new format or override the existing one. 
The templates you need are `code/lang/typeFormat/<type>/<format>/usage` and `code/lang/typeFormat/<type>/<format>/definition`,
where `<type>` is the JSON Schema type (string, integer, number, boolean, array, object) and `<format>` is the format name.
These templates must render the type definition and its usage.

For example, to add support for `integer` type with `long` format, create both templates with identical content:

```gotemplate
{{ define "code/lang/typeFormat/integer/long/definition" }}
    int64
{{- end }}

{{ define "code/lang/typeFormat/integer/long/usage" }}
    int64
{{- end }}
```

Put them in a file with any name, say `my_templates/jsonschema_formats.tmpl`, and run the code generation:

```bash
go-asyncapi code --template-dir my_templates my_asyncapi.yml
```

As a result, all JSON Schema objects with `type: integer` and `format: long` will be rendered as `int64` type in Go code.

{{% hint info %}}
In most cases, it's sufficient to leave the "usage" and "definition" templates identical, but sometimes you may want to 
make them different. For example, if a format must be represented as a struct:

```gotemplate
{{ define "code/lang/typeFormat/object/myStruct/definition" }}
    type {{goIDUpper .}} struct {
        Field1 string `json:"field1"`
        Field2 int    `json:"field2"`
    }
{{- end }}

{{ define "code/lang/typeFormat/object/myStruct/usage" }}
    {{ goPkg . }}{{ goIDUpper . }}
{{- end }}
```

See also [templating guide]({{< relref "/templating-guide/overview#definition-and-usage-code" >}}) for more details.
{{% /hint %}}
