---
title: "cp"
weight: 323
description: "Copying the nodes"
---

# Copying the nodes

## TLDR;

`go-asyncapi doc cp` copies the nodes within a document or between several documents the same way you would copy files in a Unix shell.
It supports **globbing**, **recursive copying**, **conflict resolution** (automatic and interactive), 
and automatically **keeps all affected `$ref`s correct** after the copy operation.

{{% hint info %}}
To move nodes instead of copying them, see the [`doc mv`]({{<relref "/commands/doc/mv">}}) command.
{{% /hint %}}

## Arguments

The command accepts one or more **locations** — a path to a node in [JSON Pointer]({{<relref
"/asyncapi-specification/references">}}) format, or a globbing pattern. Globbing supports `*`, `**`, etc.
(see the [glob syntax](https://github.com/gobwas/glob)). Both YAML and JSON documents are supported.

By default the **last** location is the destination and the others are the sources. A copied node is added as a
new key in a destination object, or appended as a new item to a destination array. When the destination path does
not exist, pass `-a` to create it.

## Usage

Copy active servers from one document into another:

```bash
go-asyncapi doc cp 'file1.yaml#/servers' 'file2.yaml'
```

Duplicate a single node within the same document:

```bash
go-asyncapi doc cp 'file1.yaml#/servers/foo' 'file1.yaml#/servers/bar'
```

Copy every active server into another document's `components` section (use `-a` to create it if missing):

```bash
go-asyncapi doc cp -a 'file1.yaml#/servers/*' 'file2.yaml#/components/servers'
```

Walk schema components recursively and copy every nested node whose key starts with `field`:

```bash
go-asyncapi doc cp -a 'file1.yaml#/components/schemas/**/field*' 'file2.yaml#/components/schemas'
```

Convert a YAML document to JSON:

```bash
go-asyncapi doc cp file1.yaml file2.json
```

When `-t` is given, **all** arguments are treated as sources — handy for piping a list of nodes:

```bash
cat nodes.txt | xargs go-asyncapi doc cp -a -t file2.yaml
```

## Recursive mode

By following `$ref`s, the command can also copy the nodes a node matched to **locations** depends on:

- `-r` — copy all dependencies recursively.
- `-S` — copy only the direct dependencies (resolve `$ref`s in the matched nodes, but do
  not descend further).
- `-H` — copy only the dependencies, excluding the matched nodes themselves. Useful to extract
  dependencies into a separate document. Combine it with `-r` or `-S`, otherwise nothing is copied.

{{% hint warning %}}
External and remote `$ref`s are allowed to be followed by `doc cp`.
{{% /hint %}}

### Making the inner node a separate component once referred from somewhere

On recursive copying, command turns every `$ref` that points to any inner node inside a component (`components/*` section)
or active entity (`servers`, `channels`, `operations` sections) into a separate component.
The reason is if it would copy the inner node verbatim to the same path without the parent node,
it could make the destination document invalid. 

For example, once we meet a `$ref` pointing to `#/servers/foo/bindings`, copying this node alone to `#/servers/foo/bindings`
without `#/servers/foo` contents would produce an invalid server entity because of lack of "host" and "protocol" required fields.

Instead, the command transforms `#/servers/foo/bindings` into `#/components/serverBindings/generatedName` and updates `$ref`s 
to point to the new location.

## Conflict resolution

If a node is already exists on destination path and *is not equal to source node*, the command reports a **conflict**.

By default, command exits with an error. By passing the `-f` flag, you can
instruct the command to overwrite the existing node on conflict. 
If you want to resolve each conflict interactively, pass `-i`.

## Rewriting the $refs

After copying, the command modifies `$ref`s so they remain correct in every affected document. 

{{< figure src="images/doc-cp.svg" alt="doc cp copying a node and rewriting $refs" >}}

For example, if `dest.yaml` held a `$ref` to a node in `source.yaml` that was just copied into `dest.yaml`, 
that `$ref` becomes local and vice versa. `$ref`s in unrelated `file3.yaml` (if it was passed as a source)
that pointed at the moved node are repointed as well. 

Rewriting can be disabled by passing the `--disable-rewriting` flag to leave all `$ref`s untouched.
