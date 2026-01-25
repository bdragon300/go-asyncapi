---
title: "Add a Security type"
weight: 550
description: "Adding a new Security type or overriding an existing one"
---

# Add a Security type or override an existing one

`go-asyncapi` supports several security types out-of-the-box as listed in [features page]({{< relref "/features#security-schemes" >}}).
You can add a new security type or override an existing one by defining custom templates.

{{% hint note %}}
Unknown security types, that are not built-in nor added as described below, are ignored.
{{% /hint %}}

Let's add a new `fooAuth2` security type. For that, create a template and put them to any file, say `my_templates/fooAuth2.tmpl`:

```gotemplate
{{ define "code/security/fooAuth2" }}
func New{{goID .}}Security(credentials func() (secret string)) {{goID .}}Security {
    res := {{goID .}}Security{}
    res.credentials = credentials
    return res
}

type {{goID .}}Security struct {
    credentials func() (string)
}

func (s {{goID .}}Security) AuthType() string {
    return "fooAuth2"
}

func (s {{goID .}}Security) FooAuth2() string {
    return s.credentials()
}

{{/* Associate this security object with the servers/operations that it was bound to */}}
{{- range .BoundServers}}
    func (s {{goID $}}Security) {{. | goID}}Security() {}
{{- end}}
{{- range .BoundOperations}}
    func (s {{goID $}}Security) {{. | goID}}Security() {}
{{- end}}

{{- end}}
```

Then specify the template directory when generating the code:

```bash
go-asyncapi code --template-dir my_templates my_asyncapi.yml
```

The following security entity will be rendered as `MyAuthSecurity` struct in Go code:

```yaml
components:
  securitySchemes:
    myAuth:
      type: fooAuth2
      description: My custom fooAuth2 security scheme
```

Now, this code can be used in implementation code (see also [Use an existing library]({{< relref "/howtos/use-an-existing-library" >}})):

```go
import (
    "fmt"
    "github.com/your/repo/myprotocol"
    "github.com/asyncapi/go-asyncapi/pkg/run"
)

type MyProtocolClient struct {
	Secret string
}

type FooAuth2Security interface {
    FooAuth2() string
}

func NewClient(serverURL string, bindings *myprotocol.ServerBindings, security run.AnySecurityScheme) (*MyProtocolClient, error) {
    // ...
    fooAuthSecurity, ok := security.(FooAuth2Security)
    if !ok || security.AuthType() != "fooAuth2" {
        return nil, fmt.Errorf("expected FooAuth2, got %v", security.AuthType())
    }
    token := fooAuthSecurity.FooAuth2()
    // use the token ...
    // ...
    return &MyProtocolClient{Secret: token}, nil
}
```