---
title: "Add a content type"
weight: 550
description: "Adding a new content type or overriding an existing one"
---

# Add a content type or override an existing one

Content type is MIME type, that specifies how the data is encoded in the message payload. Content types supported by 
`go-asyncapi` out-of-the-box are listed in [features page]({{< relref "/features#content-types" >}}).

{{% hint note %}}
Unknown content types, that are not built-in not added as described below, are treated as `application/json`.
{{% /hint %}}

Content type-related code is generated inside `Marshal<Protocol>` and `Unmarshal<Protocol>` methods of the message types.
So it should marshal and unmarshal the message payload to/from method's parameter.

The following example shows how to add the code for content type `text/plain`. 

For that define two following templates and put them to any file, say `my_templates/content_types.tmpl`:

```gotemplate
{{ define "code/proto/mime/messageDecoder/text/plain" }}
    var b {{ goPkgExt "strings" }}Builder
    if _, err := {{ goPkgExt "io" }}Copy(&b, r); err != nil {
        return {{ goPkgExt "fmt" }}Errorf("read message: %w", err)
    }
    m.payload = {{goPkg .PayloadType}}{{.PayloadType | goID}}(b.String())
{{- end }}

{{ define "code/proto/mime/messageEncoder/text/plain" }}
    r := {{ goPkgExt "strings" }}NewReader(string(m.Payload))
    if _, err := {{ goPkgExt "io" }}Copy(w, r); err != nil {
        return {{ goPkgExt "fmt" }}Errorf("write message: %w", err)
    }
{{- end }}
```

Then specify the template directory when generating the code:

```shell
go-asyncapi code -T my_templates my_asyncapi.yaml
```

Couple of notes:

* Template functions `goPkgExt` and `goPkg` are used to correctly import a package (or package where a type is defined)
  and get the imported prefix. For more details see [templating guide]({{< relref "/templating-guide/functions" >}}).
* Replacing the existing content type is done the same way.
