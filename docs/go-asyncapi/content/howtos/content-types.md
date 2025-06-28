---
title: "Content types"
weight: 720
description: "Supporting a new content type in the generated code"
---

# Content types

Content type is MIME type, that specifies how the data is encoded in the message payload. 

`go-asyncapi` fully supports the content type property, both a message-specific and default document-wide.

Content types supported by `go-asyncapi` are listed in [features page]({{< relref "/features#content-types" >}}).

## Default behavior

`go-asyncapi` generates the encoder/decoder code depending on the content type set in AsyncAPI message. If it's missed, 
the `defaultContentType` document field is used. If it's missed as well, the default content type is `application/json`.

To render the marshalling/unmarshalling code, `go-asyncapi` looks for templates `code/proto/mime/messageDecoder/<mime>`
and `code/proto/mime/messageEncoder/<mime>`. `mime` is the content type name (MIME type),
for example, `application/json`, `text/plain`, etc.

If not found, the default templates are `code/proto/mime/messageDecoder/default` and 
`code/proto/mime/messageEncoder/default` (which are also may be customized).

This default behavior can be customized in templates `code/proto/message/encoder` and `code/proto/message/decoder`.

[More about templates customization]({{< relref "/templating-guide/overview" >}}).

## Adding a new content type

It is very easy, just define templates `code/proto/mime/messageDecoder/<mime>` and `code/proto/mime/messageEncoder/<mime>`.
The result will be substituted to `Marshal` and `Unmarshal` message methods instead of default encoder/decoder.

For example, to add support for `text/plain` content type, define the templates `code/proto/mime/messageDecoder/text/plain`
and `code/proto/mime/messageEncoder/text/plain` with encoding/decoding code and put them to a file with any name, 
say `content_types.tmpl`.

For example:

```gotemplate
{{ define "code/proto/mime/messageDecoder/text/plain" }}
var b {{ goQual "strings.Builder" }}
if err := {{ goQual "io.Copy" }}(&b, envelope); err != nil {
    return nil, fmt.Errorf("read message: %w", err)
}
m.Payload = b.String()
{{ end }}

{{ define "code/proto/mime/messageEncoder/text/plain" }}
r := {{ goQual "strings.NewReader" }}(m.Payload)
if _, err := {{ goQual "io.Copy" }}(envelope, r); err != nil {
    return fmt.Errorf("write message: %w", err)
}
{{ end }}
```

Then run the code generation:

```shell
go-asyncapi code -T my_templates my_asyncapi.yaml
```

## Replacing the default encoder/decoder

Everything the same as for adding a new content type, i.e. by defining the templates
`code/proto/mime/messageDecoder/<mime>` and `code/proto/mime/messageEncoder/<mime>`.
See [Adding a new content type](#adding-a-new-content-type) section above.