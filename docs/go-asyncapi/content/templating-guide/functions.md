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
func goLit(value any) (string, error)
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

The qualified name is the form of writing the Go identifier, that we want to import and use in-place in the generated code.

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
func goDef(r common.GolangType) (string, error)
```

Returns the Go *definition code* for the given type, and additionally pins it to the current package (if it's pinnable).
[More about pinning]({{<relref "/templating-guide/overview#pinning">}}).

See also `pin` function, which pins any pinnable artifact without producing any code.

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
func goUsage(r common.GolangType) (string, error)
```

Returns the Go *usage code* for the given Go type artifact, automatically adding the necessary package imports if needed.

The function *may* return error `context.ErrNotPinned` in case when `r` or its dependencies 
(e.g. a field of inline struct `r`) are not pinned yet, so it can't make a proper import. 
In this case, you must ensure that both `r` and the types it consists (if any) of are all has been pinned before. 
[More about pinning]({{<relref "/templating-guide/overview#pinning">}}).

Examples:

`{{ goUsage .Type }}` if `.Type` is `MySchema` struct pinned to the current package, the function returns `MySchema`.

`{{ goUsage .Type }}` if `.Type` is `MySchema` struct in package "foo", the function returns `foo.MySchema`, adding an
import for "foo" package if necessary.

`{{ goUsage .Type }}` if `.Type` is pointer to the inline anonymous struct, the result is:

```go
*struct {
    Foo string
}
```

### goPkg

```go
func goPkg(obj any) (string, error)
```

Returns the package path prefix (name + `.`) where an artifact was pinned in the generated code (by `pin` or `goDef` functions), 
also adding it to the current file imports if necessary. If type is pinned to current package, returns empty string.

If `obj` was not pinned before, the function returns error `context.ErrNotPinned`, 
which means you must ensure that `obj` is pinned to one of generated packages before calling this function.
[More about pinning]({{<relref "/templating-guide/overview#pinning">}}).

If `obj` is `*common.ImplementationObject` the returned value is the package path to implementation code. In this case,
pinning is not needed.

Examples:

`{{ goPkg .Type }}{{ goUsage .Type }}` if `.Type` is `MySchema` is pinned to `schemas/my_schema.go`, 
and the current file is `channels/my_channel.go`, then the produced code is `schemas.MySchema`.

`{{ goPkg .Type }}{{ goUsage .Type }}` if `.Type` is `MySchema` is pinned to a file in the current package, 
then the produced code is just `MySchema` (`goPkg` returned an empty string).

## Artifact helpers

These functions are helper functions to work with artifacts. They don't produce any Go code.

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
func isVisible(a common.Artifact) common.Artifact
```

Returns `nil` if the given artifact should NOT be rendered in the generated code due to `go-asyncapi` configuration, 
`x-ignore` flag set in the AsyncAPI entity definition or other reasons.

Example:

`{{ with isVisible .Type }}{{ goDef . }}{{ end }}` produces the type definition only if `.Type` is visible,
otherwise skips the body of the "with" statement due to `nil` function result.

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

`{{ tmpl (print "code/proto/" $protocol "channel/newFunction/block1") $ }}` 
executes the `code/proto/kafka/channel/newFunction/block1` template if `$protocol == "kafka"`, 
passing the current template context (i.e. `$`) to it.

### tryTmpl

```go
func tryTmpl(templateName string, ctx any) (string, error)
```

The same as `tmpl`, but if the template is not found, returns an empty string instead of raising an error.
Useful when you want to execute a template only if it exists.

Example:

```gotemplate
{{ with tryTmpl (print "code/proto/" $protocol "channel/newFunction/block1") $ }}
    Done
{{ end }}
```

The snippet above produces the "Done" text after the code produced by `code/proto/kafka/channel/newFunction/block1`
template for `$protocol == "kafka"` only if this template exists. Otherwise, it skips the "with" body.

## Template namespace

The functions described in this section work with the template namespace.

### pin

```go
func pin(a common.Artifact) (string, error)
```

Pins (i.e. associates) the given artifact in the current rendered go package. If the artifact is already pinned, does nothing.
If the artifact cannot be pinned because it can't be declared (e.g. built-in Go type), returns an error. 
After the pinning, the given artifact become available to be imported using `goPkg` or `goUsage` functions.
[More about pinning]({{<relref "/templating-guide/overview#pinning">}}).

See also `goDef` function, which pins an artifact as well.

Returns empty string.

Example:

`{{ pin . }}` pins the `.` artifact to the current rendered Go package.

### once

```go
func once(r any) any
```

Accepts any comparable object and returns it back only once per all template executions, 
all next calls with the same argument return `nil`.

Useful to avoid duplicate code generation. The functionality is similar to `sync.Once` in Go.

Example:

```gotemplate
{{ with once .FooBar }}
    {{ goDef . }}
{{ end }}
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

### ellipsisStart

```go
func ellipsisStart(maxLen int, s string) string
```

The function truncates a string to a `maxLen` characters from the end and prepends an ellipsis ("...") if the 
string exceeds that width.

Example:

`{{ "Hello, World!" | ellipsisStart 10 }}` returns `... World!`.

### debug

```go
func debug(args ...any) string
```

Prints the given arguments to the logging output with the `debug` level and returns an empty string.

### runtimeExpressionCode

```go
func runtimeExpressionCode(c lang.BaseRuntimeExpression, target *lang.GoStruct, addValidationCode bool) ([]runtimeExpressionCodeStep, error)
```

Special purpose function that accepts the struct and runtime expression and returns the Go code that extracts the 
value from the struct according to the runtime expression.

Parameters:
1. `lang.BaseRuntimeExpression` -- object with runtime expression info. Every artifact that has
   a runtime expression (e.g. `lang.CorrelationID`) contains a field of this type.
2. `lang.GoStruct` -- target struct where the value should be extracted from.
3. `addValidationCode` -- if `true`, the result also contains the additional error handing code,
   that is typically used for property getter functions.
   
Example:

```gotemplate
v0 := m.{{.CorrelationID.StructFieldKind | toString | untitle}}
{{- $steps := runtimeExpressionCode .RuntimeExpression .TargetType true }}
{{- range $steps }}
    {{- range .CodeLines}}
        {{.}}
    {{- end}}
{{- end }}

{{if $steps}}value = {{last $steps | .VarName}}{{end}}
```
