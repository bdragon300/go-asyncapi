---
title: "tree"
weight: 327
description: "Showing the documents linked by $refs"
---

# Documents tree

`go-asyncapi doc tree` shows the hierarchy of documents linked by `$ref`s, starting from the given document.
It is handy to get an overview of how a spec is split across several files and which documents depend on which.

## Arguments

Command receives one argument, the path to the document to start from. The path can be a local file or a URL.

By default the hierarchy is drawn as a graphical **tree**. The `-l` arg turns the output into a plain
text list, one document per line, which is suitable for scripting.

By default the `$ref`s pointing to URLs are not followe, pass `-F` to allow following them.

{{% hint tip %}}
The `$ref` locator options are described in the
[reference locator]({{<relref "/asyncapi-specification/references#reference-locator">}}) article.
{{% /hint %}}

## Usage

Show the documents tree starting from a document:

```bash
go-asyncapi doc tree asyncapi.yaml
```

Print the documents as a plain list, one per line:

```bash
go-asyncapi doc tree asyncapi.yaml -l
```

Follow the `$ref`s that point to URLs as well:

```bash
go-asyncapi doc tree asyncapi.yaml -F
```
