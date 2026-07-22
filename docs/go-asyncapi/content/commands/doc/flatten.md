---
title: "flatten"
weight: 326
description: "Flattening the AsyncAPI documents"
---

# Flattening the AsyncAPI documents

## TLDR;

`go-asyncapi doc flatten` command resolves all `$ref`s in a document and replaces them with the objects they refer to,
producing a single self-contained document.

## Arguments

Command receives one argument, the path to the document to flatten. 
The path can be a local file or a URL, both YAML and JSON documents are supported.

By default, only `$ref`s to local objects (inside the same document) are replaced. To also resolve `$ref`s
to other documents on the local filesystem, use `--external-refs`. To resolve `$ref`s addressed by HTTP(s)
URLs, use `--remote-refs`.

{{% hint info %}}
The AsyncAPI specification explicitly demands `$ref` to be used in certain places, e.g. `channel` field in "Operation Object", etc.
`flatten` respects this and leaves these `$ref`s unresolved to keep the document valid.

See [AsyncAPI specification](https://www.asyncapi.com/docs/reference/specification/v3.1.0).
{{% /hint %}}

{{% hint tip %}}
The `$ref` locator options are described in the
[reference locator]({{<relref "/asyncapi-specification/references#reference-locator">}}) article.
{{% /hint %}}

## Usage

To flatten a document in-place:

```bash
go-asyncapi doc flatten my-app.yaml
```

To write the result to another file instead of modifying the original, use the `-o` (`--output`) option:

```bash
go-asyncapi doc flatten my-app.yaml -o flattened.yaml
```

To inline `$ref`s pointing to other documents on the local filesystem, use `--external-refs`:

```bash
go-asyncapi doc flatten --external-refs my-app.yaml
```

To inline all `$ref`s, including those addressed by URLs, add `--remote-refs`:

```bash
go-asyncapi doc flatten --external-refs --remote-refs my-app.yaml
```
