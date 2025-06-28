---
title: "Internals"
weight: 1000
description: "How the go-asyncapi works internally"
---

# go-asyncapi internals

This article briefly describes the internal structure of the `go-asyncapi`.

The figure below shows the execution steps that `go-asyncapi` performs to produce the output.

{{< figure src="images/internals.svg" alt="Execution steps" >}}

### 1. Compilation

Compilation step contains the logic to parse and convert an AsyncAPI entity to the object form, suitable for rendering. The next
steps work only with these objects. For example, it's on this step a jsonschema object is turned into the Go struct (in object form).

On this step all entities are compiled into the internal representation objects called *artifacts*.

Every artifact satisfies the `common.Artifact` interface. 
Some artifacts represent the complex entities (e.g. a channel) and produce the complex code.
Others are simpler and represent a simple Go type (e.g. a jsonschema object), they additionally satisfy the 
`common.GolangType` interface.

The result of the compilation step is a list of artifacts, gathered and compiled from a passed AsyncAPI document
and all referenced documents.

#### Locator

If an AsyncAPI document contains one or more `$ref` to another document, `go-asyncapi` uses the 
[Locator]({{< relref "/asyncapi-specification/references" >}}). This is a part of `go-asyncapi`, 
that locates and reads a document by its URL using either the built-in logic or user-provided command.

`go-asyncapi` collects external `$ref` from the given document and feeds them to the Locator, if any. 
Locator in turn gets a document and sends it to the compiler back. 
Compiler handles another document, collecting more external `$ref`s and passing them to the Locator again.
This process continues until all documents are fetched.

#### Late binding

Just like entities in the AsyncAPI document, artifacts are also linked to each other. For example, a channel refers 
to a message by `$ref`, a message includes the jsonschema definition in payload section, and so on. 

To manage this we use the "late binding" technique. Placeholder types `lang.Ref` and `lang.Promise` (blue squares on the
figure above) keep the reference to an artifact, that will appear after the linking step.

We use the late binding instead of handling everything immediately on compilation step by a couple of reasons.

The first reason is that it simplifies the tool design keeping in mind our list of supported features. 
Theoretically, we may get a document contained the long chains of references, recursive references, and so on.
These chains may also contain references to external documents, and entities there may in turn to refer to 
another document or even back to the current document. 

If we handle all of these cases immediately, we could get the dependency hell. Instead, in `go-asyncapi` we compile all 
entities into artifacts independently as a separate step, and then use the late binding to link them together without 
long recursive calls (linking step).

Another reason is that it would be better to ignore the order of entity definitions in document -- changing the order
of server definitions should not affect the result.

As a drawback of this approach, this requires another execution step, and we have `lang.Promise` and `lang.Ref` types 
in the code.

### 2. Linking

On the linking step, `go-asyncapi` does the late binding process described above. Specifically, it walks through
all `lang.Ref` and `lang.Promise` and fills them with the pointers to artifacts they refer to. 
If the artifact is not found (e.g. due to incorrect `$ref`), it raises an error.

The result of the linking step is the same list of artifacts, but with all references resolved.

### 3. Rendering

The rendering step is the most important one. It takes the list of artifacts from the previous step and renders them 
into the selected output format using the [Go templates](https://pkg.go.dev/text/template).

Basically, the rendering step is a loop over all artifacts, where every artifact is rendered using the root template, 
passing an artifact as a context defined in `tmpl` package. The result is merged into the output file(s).

In code generation mode, every artifact additionally are running through the 
[code layout]({{< relref "/howtos/code-layout" >}}) rules and rendered for every matching rule, 
merging the result in layout file structure.
After the process is finished, every resulting file is additionally processed by the preamble template, 
that is used to add the package declaration, import statements, "copyright" notice, etc.

### 4. Formatting

This post-processing step is optional and is used to format the result. For Go code, the `gofmt` tool is used to format the code.

### 5. Writing

The final step writes the results to files.
