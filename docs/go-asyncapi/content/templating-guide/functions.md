---
title: "Functions reference"
weight: 920
bookToC: true
description: "Template functions reference"
---

# Functions

[Template functions](https://pkg.go.dev/text/template#hdr-Functions) are used to call the Go functions from the templates
and use their results in the template rendering.

The biggest part of functions are provided by [sprout](https://github.com/go-sprout/sprout) package. There are a lot
of functions to work with types, numbers, environment variables, filesystem, etc. 

[Full list of sprout functions](https://docs.atom.codes/sprout/registries/list-of-all-registries).

Another part of functions is provided by the `go-asyncapi` itself and have functionality specific to the `go-asyncapi` tool.
Sections below describe the functions provided by `go-asyncapi`.

## Code generation

The functions described in this section have the `go` name prefix, which means they return the Go code snippets.

### goLit

```go
func goLit(value any) string
```

Converts the given value to a Go literal.

If value is one of built-in Go types, it returns the value as a Go literal.
If value is a `common.GolangType`, it returns the Go *usage code* of the type (equivalent of `goUsage(value.(common.GolangType))`).

Examples:

`{{ goLit 42 }}` returns `42`.

`{{ goLit "foo" }}` returns `"foo"`.

`{{ goUsage .Type }}` if `.Type` is pointer to `MySchema` struct, the function returns `*MySchema`.

### goID

```go
func goID(value any) string
```

Returns the *unexported* Go identifier for the given type. 

For `common.Artifact` value the function returns the artifact's Go identifier (type name), 
for `string` value, it converts the string to a valid Go identifier. Otherwise, the function panics.

Examples:

`{{ goID .Type }}` if `.Type` is pointer to `MySchema` struct, the function returns `mySchema`.

`{{ goID "foo_bar http" }}` returns `fooBarHTTP`.

### goIDUpper

```go
func goIDUpper(value any) string
```

Returns the *exported* Go identifier for the given type.

For `common.Artifact` value the function returns the artifact's Go identifier (type name),
for `string` value, it converts the string to a valid Go identifier. Otherwise, the function panics.

Examples:

`{{ goIDUpper .Type }}` if `.Type` is pointer to `MySchema` struct, the function returns `MySchema`.

`{{ goIDUpper "foo_bar http" }}` returns `FooBarHTTP`.

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

### goQual

```go
func goQual(parts ...string) string
```

Joins the parts into a qualified Go name and returns it, also adding it to the current file imports if necessary.

The qualified name is the form of writing the Go identifier,
that we want to import and use in-place in the generated code.

Syntax of the qualified name is as follows:

```
[<package name|module name>.]<identifier>
```

Examples:

`var foo {{ goQual "Foo" }}` produces `var foo Foo`.

`var ctx {{ goQual "context" "Context" }}`, `var ctx {{ goQual "context.Context" }}` are equivalent. 
Produces the `var ctx context.Context` and adds the `context` package to file imports.

`var foo {{ goQual "golang.org/x/net/ipv4" "GolangType" }}`, `var foo {{ goQual "golang.org" "x/net/ipv4" "GolangType" }}`,
`var foo {{ goQual "golang.org/x/net/ipv4.GolangType" }}` are equivalent. 
Produces `var foo common.GolangType` and adds the `golang.org/x/net/ipv4` package to file imports.

### goQualR

```go
func goQualR(parts ...string) string
```

The same as `goQual`, but it additionally appends the `go-asyncapi`'s "runtime" module path before the parts.

"Runtime" module path is configurable, by default it is `github.com/bdragon300/go-asyncapi/run`.
See also [Configuration reference]({{< relref "/configuration" >}}).

Examples:

`var foo {{ goQualR "ParamString" }}` produces `var foo run.ParamString` and adds the `github.com/bdragon300/go-asyncapi/run` package to file imports.

`var foo {{ goQualR "kafka" "ServerBindings" }}`, `var foo {{ goQualR "kafka.ServerBindings" }}` are equivalent.
They produce the `var foo kafka.ServerBindings` and adds the `github.com/bdragon300/go-asyncapi/run/kafka` package to file imports.

### goDef

```go
func goDef(r common.GolangType) string
```

Returns the Go *definition code* for the given type, also adding it to the current file's definitions list.

Examples:

`{{ goDef .Type }}` if `.Type` is `MySchema` struct, the function may return

```go
type MySchema struct {
    Foo string
}
```

`{{ goDef .Type }}` if `.Type` is type alias `Foo` to `int`, the function returns `type Foo int`.

### goUsage

```go
func goUsage(r common.GolangType) string
```

Returns the Go *usage code* for the given Go type artifact.

Examples:

`{{ goUsage .Type }}` if `.Type` is `MySchema` struct, the function returns `MySchema`.

`{{ goUsage .Type }}` if `.Type` is pointer to the inline anonymous struct, the function may return

```go
*struct {
    Foo string
}
```

### goPkg

```go
func goPkg(obj any) string
```

Returns the package path prefix where a type is declared in the generated code, also adding it to the current file 
imports if necessary. If type is defined in current package, returns empty string.

If the type is not defined anywhere before (`goDef`, `def` functions), raises error `context.ErrNotDefined`.

For `common.GolangType` value the result is the package path where type is declared.
For the `*common.ImplementationObject` the returned value is the package path of the given Implementation code. 
Otherwise, the function raises error.

Examples:

`{{ goPkg .Type }}{{ goUsage .Type }}` if `.Type` is `MySchema` is declared in `schemas/my_schema.go`, 
the current file is `channels/my_channel.go`, then the produced code is `schemas.MySchema`.

`{{ goPkg .Type }}{{ goUsage .Type }}` if `.Type` is `MySchema` is declared in the same package, 
then the produced code is `MySchema`.

## Artifact helpers

These functions are helper functions to work with artifacts. They don't produce any Go code.

### deref

```go
func deref(r common.Artifact) common.Artifact
```

Dereferences the `lang.Ref` or `lang.Promise` object and returns the artifact they refer to. 
If the argument is not `lang.Ref` or `lang.Promise`, just returns it back.

Example:

`{{ deref . | goUsage }}` if `.Type` is ref to `MySchema`, the template result is `MySchema`.

### innerType

```go
func innerType(r common.GolangType) common.GolangType
```

Returns the inner type of the given *wrapper type* (pointer, type redefinition). If r is not a wrapper type, returns `nil`.

Example:

The template `{{ with innerType .Type }}{{ goUsage . }}{{ end }}` produces:

* `MySchema` if `.Type` is pointer type `*MySchema`
* `int` if `.Type` is a `Foo` type, that redefines `int` (`type Foo int`)
* Nothing, if `.Type` is something else, "with" body is skipped due to `nil` result.

### isVisible

```go
func isVisible(r common.Artifact) common.Artifact
```

Returns `nil` if the given artifact should NOT be rendered in the generated code due to `go-asyncapi` configuration, 
`x-ignore` flag set in the AsyncAPI entity definition or other reasons.

Example:

`{{ with isVisible .Type }}{{ goDef . }}{{ end }}` produces the type definition only if `.Type` is visible,
otherwise skips the body of the "with" statement due to `nil` function result.

### ptr

```go
func ptr(r common.GolangType) common.GolangType
```

Wraps the r with `lang.GoPointer` and returns it. If r is `nil`, raises error.

Examples:

`{{ goUsage (ptr .Type) }}` if `.Type` is `MySchema`, the template result is `*MySchema`.

`{{ goUsage (ptr (ptr .Type)) }}` if `.Type` is `MySchema`, the template result is `**MySchema`.

## Template execution

Although the Go template language has the `template` directive that executes another template, it the compile-time directive,
so it doesn't support the dynamic template names.

The functions described in this section are used to execute the templates dynamically, at runtime.

### tmpl

```go
func (templateName string, ctx any) (string, error)
```

Looks up the template with the given name and executes it passing the `ctx` as a context object.
If the template is not found or template execution fails, returns an error.

Example:

`{{ tmpl (print "code/proto/" $protocol "channel/newFunction/block1") $ }}` 
executes the `code/proto/kafka/channel/newFunction/block1` template if `$protocol == "kafka"`, 
passing the current template context (i.e. `$`) to it.

### tryTmpl

```go
func tryTmpl(templateName string, ctx any) string
```

The same as `tmpl`, but if the template is not found, returns an empty string instead of raising an error.
Useful when you want to execute a template only if it exists.

Example:

```gotemplate
{{ with (tryTmpl (print "code/proto/" $protocol "channel/newFunction/block1") $) }}
    Done
{{ end }}
```

The snippet above produces the "Done" text after the code produced by `code/proto/kafka/channel/newFunction/block1`
template for `$protocol == "kafka"` only if this template exists. Otherwise, it skips the "with" body due to empty string
result of `tryTmpl`.

## Template namespace

The functions described in this section work with the template namespace.

### def

```go
func def(objects ...any) string
```

Explicitly define the given value(s) in template namespace.

If an object is string, it will be just added to the namespace.

If an object is `common.GolangType`, then it will be marked as declared in the current file.
Actually, the only use case of this option is to be able to render the *usage code* for the type before its *definition code*
has been rendered (i.e. before `goDef` function is called for this type).

[defined](#defined) function returns `false` for the type that is defined with `def` function, but not rendered yet. 

### defined

```go
func defined(r any) bool
```

Returns `true` if the given value is defined in the current template namespace.

If the value is string, it checks if the string is defined in the namespace.

If the value is `common.GolangType`, it checks if the type's declaration has been rendered by `goDef` function.

Example:

The template

```gotemplate
{{ if defined "foobar" }}foobar is defined{{ else }}foobar is NOT defined{{ end }}
{{def "foobar"}}
{{ if defined "foobar" }}foobar is defined{{ else }}foobar is NOT defined{{ end }}
```

produces the following output:

```
foobar is NOT defined
foobar is defined
```

### ndefined

```go
func ndefined(r any) bool
```

Opposite of the `defined` function. Returns `true` if the given value is NOT defined in the current template namespace.

Examples:

`{{ if ndefined .Type }}{{ goDef . }}{{ end }}` produces only one type definition not matter how many times the template is executed.

The template

```gotemplate
{{ if ndefined "foobar" }}foobar is NOT defined{{ else }}foobar is defined{{ end }}
{{def "foobar"}}
{{ if ndefined "foobar" }}foobar is NOT defined{{ else }}foobar is defined{{ end }}
```

produces the following output:

```
foobar is NOT defined
foobar is defined
```

## Other helpers

### impl

```go
func impl(protocol string) *common.ImplementationObject
```

Returns the `common.ImplementationObject` for the given protocol. If the protocol is not supported or the implementation
for this protocol is disabled in configuration, returns `nil`.

Example:

The template `{{ with impl "kafka" }}{{ .Manifest.Name }}{{ end }}` produces the selected Kafka implementation name or
skips the "with" body if Kafka implementation is disabled.

### toQuotable

```go
func toQuotable(s string) string
```

Returns the quotable literal string for the given s, which can be safely wrapped in double quotes in the generated Go code.

The function escapes the double quotes, escape sequences, control characters, non-printable characters.

Example:

`{{ toQuotable "foo\n \"bar\" \xFFbaz" }}` returns `foo\n \"bar\" \xFFbaz`.

### debug

```go
func debug(args ...any) string
```

Prints the given arguments to the logging output with the `debug` level and returns an empty string.

### correlationIDExtractionCode

```go
func correlationIDExtractionCode(c *render.CorrelationID, varStruct *lang.GoStruct, addValidationCode bool) (items []correlationIDExtractionStep, err error)
```

Special purpose function that generates and returns the list of steps in the generated code to extract the 
correlation ID from the message struct.

Parameters:
1. `render.CorrelationID` -- correlation ID artifact
2. `lang.GoStruct` Go struct artifact, that contains the field, or where a nested struct contains the field, 
   which the correlation ID is pointed to.
3. if `addValidationCode` is `true`, the function additionally inserts the code to check if the variable passed to the
   correlation ID getter function has the value the correlation ID points to. 
   If not, this code sets the `err` variable to an error.
   
Example:

```gotemplate
v0 := m.{{.CorrelationID.StructFieldKind | untitle}}
{{ $steps := correlationIDExtractionCode .CorrelationID .VarStruct true }}
{{ range $steps }}
    {{range .CodeLines}}{{.}}{{end}}
{{ end }}

{{if $steps}}value = {{last $steps | .VarName}}{{end}}
```
