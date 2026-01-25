---
title: "Functions reference"
weight: 920
bookToC: true
description: "Template functions reference"
---

# Functions

[Template functions](https://pkg.go.dev/text/template#hdr-Functions) are used to call the Go functions from the templates
and use their results in the template rendering.

Besides the [standard functions](https://pkg.go.dev/text/template#hdr-Functions), the biggest part of functions are 
provided by [sprout](https://github.com/go-sprout/sprout) package. 
There are a lot of functions to work with types, numbers, environment variables, filesystem, etc. 

[Full list of sprout functions](https://docs.atom.codes/sprout/registries/list-of-all-registries).

Another part of functions is provided by the `go-asyncapi` itself and have functionality specific to the `go-asyncapi` tool.
Sections below describe the functions provided by `go-asyncapi`.

## Code generation

The functions described in this sect6ion have the `go` name prefix, which means they return the Go code snippets.

### goLit

```go
func goLit(value any) (string, error)
```

Converts the given value to a Go literal.

If value is a `common.GolangType`, it returns its *usage code* -- equivalent of `{{ goUsage value }}`.
If value is one of primitive Go types, it returns the value as a Go literal. Otherwise, function panics.

Examples:

{{% hint default %}}
`{{ goLit 42 }}` returns `42`.
{{% /hint %}}

{{% hint default %}}
`{{ goLit "foo" }}` returns `"foo"`.
{{% /hint %}}

{{% hint default %}}
If `.Type` is pointer to `MySchema` struct, then `{{ goLit .Type }}` produces `*MySchema`.
{{% /hint %}}

### goIDLower

```go
func goIDLower(value any) string
```

Converts value to a correct Go identifier and returns it in *unexported* form, i.e. starting from lowercase letter. 

For `common.Artifact` value the function uses the artifact's Name (calling its `.Name` method).
If value is string, function converts it to a valid Go identifier. For other argument types the function panics.

Examples:

{{% hint default %}}
If `.Type` is pointer to `MySchema` struct, then `{{ goIDLower .Type }}` produces `mySchema`.
{{% /hint %}}

{{% hint default %}}
`{{ goIDLower "foo_bar http" }}` returns `fooBarHTTP`.
{{% /hint %}}

### goID

```go
func goID(value any) string
```

Converts value to a correct Go identifier and returns it in *exported* form, i.e. starting from uppercase letter.

For `common.Artifact` value the function returns the artifact's Name (calling its `.Name` method).
If value is string, function converts it to a valid Go identifier. For other argument types the function panics.

Examples:

{{% hint default %}}
If `.Type` is pointer to `MySchema` struct, then `{{ goID .Type }}` produces `MySchema`.
{{% /hint %}}

{{% hint default %}}
`{{ goID "foo_bar http" }}` returns `FooBarHTTP`.
{{% /hint %}}

### goComment

```go
func goComment(text string) string
```

Converts the given text to a Go comment. If text contains newlines, it is converted to a multi-line comment.

Examples:

`{{ goComment "This is a comment" }}` returns `// This is a comment`.

`{{ goComment "This is a\nmulti-line comment" }}` returns

```go
/*
This is a
multi-line comment */
```

### goPkg

```go
func goPkg(obj common.Artifact) (string, error)
```

Consume a name from any generated package seamlessly.

* If `obj` is [pinned]({{<relref "/templating-guide/overview#pinning">}}) to a package other than the current one, function 
adds an import to current file if necessary and returns the imported name prefix -- package_name+'.'.
* If `obj` is [pinned]({{<relref "/templating-guide/overview#pinning">}}) to current package, function does nothing and 
  returns the empty string.
* If `obj` not pinned at all, the function returns error `tmpl.ErrNotPinned`.

Examples:

{{% hint default %}}
If `.Type` is defined in `schemas/my_schema.go` and the current file is `channels/my_channel.go`, then
`{{ goPkg .Type }}MySchema` produces `schemas.MySchema` adding an import `import schemas`.
{{% /hint %}}

{{% hint default %}}
If `.Type` is defined in `channels/my_schema.go` (i.e. in current package `channels`) and the current file is 
`channels/my_channel.go`, then `{{ goPkg .Type }}MySchema` produces `MySchema` and function does nothing.
{{% /hint %}}

### goPkgUtil

```go
func goPkgUtil(parts ...string) (string, error)
```

Consume a name from [util code]({{<relref "/commands/code#design-overview">}}) seamlessly.

Joins `parts` into the import path, adds an import to current file if necessary and returns the imported name 
prefix -- package_name+'.'. If util code was not generated for a particular protocol, returns error `tmpl.ErrNotPinned`.

Examples:

{{% hint default %}}
`{{goPkgUtil "amqp"}}ServerBindings` produces `amqp.ServerBindings` adding an import, e.g. `example.com/foo/bar/asyncapi/proto/amqp`.
{{% /hint %}}

### goPkgImpl

```go
func goPkgImpl(parts ...string) (string, error)
```

Consume a name from [implementation code]({{<relref "/commands/code#design-overview">}}) seamlessly.

Joins `parts` into the import path, adds an import to current file if necessary and returns the imported name
prefix -- package_name+'.'. If implementation code was not generated for a particular protocol, returns error `tmpl.ErrNotPinned`.

Examples:

{{% hint default %}}
`{{goPkgImpl "amqp"}}NewClient` produces `amqp.NewClient` adding an import, e.g. `example.com/foo/bar/asyncapi/proto/amqp`.
{{% /hint %}}

### goPkgRun

```go
func goPkgRun(parts ...string) (string, error)
```

Consume a name from [runtime package]({{<relref "/commands/code#runtime-package">}}) seamlessly.

Joins `parts` into the import path, adds an import to current file if necessary and returns the imported name
prefix -- package_name+'.'.

Examples:

{{% hint default %}}
`{{goPkgRun}}ParamString` produces `run.ParamString` adding an import `github.com/bdragon300/go-asyncapi/run`.
{{% /hint %}}

### goPkgExt

```go
func goPkgExt(parts ...string) (string, error)
```

Consume a name from any external package seamlessly.

Joins `parts` into the import path, adds an import to current file if necessary and returns the imported name
prefix -- package_name+'.'.

{{% hint info %}}
Because the function arguments are joined into an import part, the expressions `{{goPkgExt "golang.org/x/net/ipv4"}}`, 
`{{goPkgExt "golang.org" "x/net/ipv4"}}`, `{{goPkgExt "golang.org" "x" "net" "ipv4" }}` are equivalent.
{{% /hint %}}

Examples:

{{% hint default %}}
`{{goPkgExt "io"}}Copy` produces `io.Copy` adding an import `io`.
{{% /hint %}}

### goDef

```go
func goDef(r common.GolangType) (string, error)
```

[Pins]({{<relref "/templating-guide/overview#pinning">}}) a type to the current package and
returns the Go [definition code]({{<relref "/templating-guide/overview#definition-and-usage-code">}}). If type is not
pinnable, then it just return the definition code.

Examples:

{{% hint default %}}
If `.Type` is `lang.GoStruct` with name "MySchema" , then `{{ goDef .Type }}` produces the code:

```go
type MySchema struct {
    Foo string
}
```
{{% /hint %}}

{{% hint default %}}
If `.Type` is type alias with name "Foo" to `int` type, then `{{ goDef .Type }}` produces `type Foo int`.
{{% /hint %}}

### goUsage

```go
func goUsage(r common.GolangType) (string, error)
```

Returns the Go [usage code]({{<relref "/templating-guide/overview#definition-and-usage-code">}}) for the given 
Go type artifact, automatically adding the necessary package imports if needed.

* If `r` is [pinned]({{<relref "/templating-guide/overview#pinning">}}) to a package other than the current one, function
  adds an import to current file if necessary and returns the imported name prefix -- package_name+'.'.
* If `r` is [pinned]({{<relref "/templating-guide/overview#pinning">}}) to current package, function does nothing and
  returns the empty string.
* If `r` not pinned at all, the function returns error `tmpl.ErrNotPinned`.

Examples:

{{% hint default %}}
If `.Type` is `MySchema` struct defined in `schemas/my_schema.go` and the current file is `channels/my_channel.go`, 
then `{{ goUsage .Type }}` produces `schemas.MySchema` adding an import `import schemas`.
{{% /hint %}}

{{% hint default %}}
If `.Type` is `MySchema` struct defined in `channels/my_schema.go` (i.e. in current package `channels`) 
and the current file is `channels/my_channel.go`, then `{{ goUsage .Type }}` produces `MySchema` without adding
any imports.
{{% /hint %}}

{{% hint default %}}
If `.Type` is pointer to the inline anonymous struct (that can't be pinned, since it can't have definition),  
then `{{ goUsage .Type }}` produces the code:

```go
*struct {
    Foo string
}
```
{{% /hint %}}

{{% hint default %}}
If `.Type` is the primitive type `int`, then `{{ goUsage .Type }}` produces `int`.
{{% /hint %}}

## Codegen helpers

These functions are helper functions to work with artifacts. They don't produce any Go code.

### pin

```go
func pin(a common.Artifact) (string, error)
```

[Pins]({{<relref "/templating-guide/overview#pinning">}}) (i.e. associates) the given artifact to the current
rendered go package. If the artifact is already pinned, does nothing. If the artifact is not pinnable, returns an error.

Always returns empty string.

Example:

{{% hint default %}}
`{{ pin . }}` pins the `.` artifact to the current rendered Go package.
{{% /hint %}}

### isVisible

```go
func isVisible(a common.Artifact) common.Artifact
```

Returns `nil` if the given artifact should NOT be rendered in the generated code due to `go-asyncapi` configuration,
`x-ignore` flag set in the AsyncAPI entity definition or other reasons.

Example:

{{% hint default %}}
`{{ with isVisible .Type }}{{ goDef . }}{{ end }}` produces the type definition only if `.Type` is visible,
otherwise it skips the body of the "with" statement due to `nil` function result.
{{% /hint %}}

### once

```go
func once(r any) any
```

Accepts any [comparable](https://go.dev/blog/comparable) object and returns it back only first time during the `go-asyncapi`
execution, all next calls with the same object return `nil`.

Useful to avoid duplicate code generation. The functionality is similar to `sync.Once` in Go.

Example:

{{% hint default %}}
```gotemplate
{{ with once .FooBar }}
    {{ goDef . }}
{{ end }}
```
{{% /hint %}}

### innerType

```go
func innerType(r common.GolangType) common.GolangType
```

Returns the type that is wrapped by *wrapper type* (pointer, type redefinition). If r is not a wrapper type, returns `nil`.

Example:

{{% hint default %}}
If `.Type` is a `lang.Pointer` or `lang.GoTypeDefinition` wrapping an `lang.GoStruct` object `MySchema`, then the expression 
`{{ with innerType .Type }}{{ goUsage . }}{{ end }}` produces `MySchema` in both cases.

If `.Type` is something else, "goUsage" won't execute.
{{% /hint %}}

### impl

```go
func impl(protocol string) *tmpl.ImplementationCodeInfo
```

Returns the `tmpl.ImplementationCodeInfo` object describing the implementation code for the given protocol.
If the implementation code was not generated for this protocol, returns `nil`.

Example:

{{% hint default %}}
The template `{{ with impl "kafka" }}{{ .Protocol }}{{ end }}` produces the protocol name only if the implementation
code was generated for "kafka" protocol.
{{% /hint %}}

### utilCode

```go
func utilCode(protocol string) *common.UtilCodeInfo
```

Returns the `tmpl.UtilCodeInfo` object describing the util code for the given protocol.
If the util code was not generated for this protocol, returns `nil`.

Example:

{{% hint default %}}
The template `{{ with utilCode "kafka" }}{{ .Protocol }}{{ end }}` produces the protocol name only if the util
code was generated for "kafka" protocol.
{{% /hint %}}

### runtimeExpression

```go
func runtimeExpression(a tmpl.runtimeExpressionArtifact, target *lang.GoStruct, addValidationCode bool) *tmpl.RuntimeExpressionInfo
```

Function compiles the Go code, that recursively extracts a value based on runtime expression and target struct.

Arguments are:

1. artifact that contains the runtime expression: `CorrelationID` or `OperationReplyAddress`
2. struct artifact to extract a value from
3. flag to include the validation code such as array bound check or map key presence check, that is typically used
   for getter methods

If `a` is nil or is not visible, function returns `nil`. If error occurred during the runtime expression compilation,
function is also returns `nil`.

Typical usage:

```gotemplate
{{- with runtimeExpression .CorrelationID .OutType false}}
    func (m *{{ $.OutType | goID }}) SetCorrelationID(value {{goUsage .OutputType}}) *{{ goID $.OutType }} {
        {{.InputVar}} := m.{{toString .Expression.StructFieldKind | toTitleCase}}
        {{template "code/runtimeExpression/setterBody" .}}
        m.{{toString .Expression.StructFieldKind | toTitleCase}} = {{.InputVar}}
        return m
    }
{{- end}}

{{- with runtimeExpression .CorrelationID $.InType false}}
    func (m {{ $.InType | goID }}) CorrelationID() (value {{goUsage .OutputType}}, err error) {
        {{.InputVar}} := m.{{toString .Expression.StructFieldKind}}
        {{template "code/runtimeExpression/getterBody" .}}
        {{if .CodeSteps}}value = {{.OutputVar}}{{end}}
        return
    }
{{- end}}
```

## Template execution

Although the Go template language has the `template` directive that executes another template, it the compile-time directive,
so it doesn't support the dynamic template names.

The functions described in this section are used to execute the templates dynamically, at runtime.

### tmpl

```go
func tmpl(templateName string, ctx any) (string, error)
```

Looks up the template with the given name and executes it passing the `ctx` as a context object.
If the template is not found or template execution fails, returns an error.

Example:

{{% hint default %}}
`{{ tmpl (print "code/proto/" $protocol "channel/newFunction/block1") $ }}` 
executes the `code/proto/kafka/channel/newFunction/block1` template if `$protocol == "kafka"`, 
passing the current template context (i.e. `$`) to it.
{{% /hint %}}

### tryTmpl

```go
func tryTmpl(templateName string, ctx any) (string, error)
```

The same as `tmpl`, but if the template is not found, returns an empty string instead of raising an error.
Useful when you want to execute a template only if it exists.

Example:

{{% hint default %}}
```gotemplate
{{ with tryTmpl (print "code/proto/" $protocol "channel/newFunction/block1") $ }}
    {{.}}
{{ end }}
```

This produces the output of the template `code/proto/kafka/channel/newFunction/block1`
for `$protocol == "kafka"` only if this template exists. Otherwise, this snippet does nothing.
{{% /hint %}}

## Other helpers

### toQuotable

```go
func toQuotable(s string) string
```

Returns the quotable string for the given s, which can be safely wrapped in double quotes in the generated Go code.

The function escapes the double quotes, escape sequences, control characters, non-printable characters.

Example:

{{% hint default %}}
`{{ toQuotable "foo\n \"bar\" \xFFbaz" }}` produces `foo\n \"bar\" \xFFbaz`.
{{% /hint %}}

### ellipsisStart

```go
func ellipsisStart(maxLen int, s string) string
```

The function truncates a string to a `maxLen` characters from the end and prepends an ellipsis ("...") if the 
string exceeds that width.

Example:

{{% hint default %}}
`{{ "Hello, World!" | ellipsisStart 10 }}` returns `... World!`.
{{% /hint %}}

### debug

```go
func debug(args ...any) string
```

Prints the given arguments to the logging output with the `debug` level and returns an empty string.

### mapping

```go
mapping(v any, variantPairs ...any) any
```

Looks for a key equal to `v` in key-value pairs and returns the associated value or `nil` if key not found.

Example:

{{% hint default %}}
`{{ mapping "spam" "foo" "bar" "spam" "eggs" 123 456 }}` produces `eggs`
{{% /hint %}}

{{% hint default %}}
`{{ with mapping "?" "foo" "bar" "spam" "eggs"}}{{.}}{{else}}Nothing{{end}}` produces `Nothing`
{{% /hint %}}

### toList

```go
toList(v any) ([]any, error)
```

Converts any array/slice `v` to `[]any` slice. Returns error if `v` is not array or slice.

### hasKey

```go
hasKey(key any, m any) (bool, error)
```

Returns `true` if `key` is present in map `m`. Return error if m is not a map or `key` type is not compatible to map's key type.
