---
title: "gen-examples"
weight: 325
description: "Generating synthetic examples"
---

# Generating the examples

## TLDR;

`go-asyncapi doc gen-examples` generates synthetic examples for the messages and schemas found in a document.

{{% hint warning %}}
The command does **not** yet support message traits, and it does **not** respect the JSON Schema value restrictions 
(`minimum`/`maximum`, `enum`, `pattern`, etc.). 
The produced values are random yet and only match the declared `type` and `format`.
{{% /hint %}}

## Arguments

The command takes a **location** — a document file or URL, optionally with a JSON pointer to a target
subtree. Both YAML and JSON documents are supported. 
If the location contains a pointer, the generation is limited to that node only. 

By default, the command modifies the supplied file in-place. Pass `-o` to write the result to another file instead.

By default, the command generates **one example** for every message and schema **without examples**. 
Add `--append` flag to generate examples for messages and schemas regardless of whether they already have examples.
Add `--only-messages` or `--only-schemas` to limit the generation to messages or schemas only.
By default, one example is generated per entity, but you can change that with the `--count` flag.

Fetching URLs in `$ref`s are denied by default for security reasons, pass `--allow-remote-refs` to allow fetching them.

{{% hint tip %}}
The `$ref` locator options are described in the
[reference locator]({{<relref "/asyncapi-specification/references#reference-locator">}}) article.
{{% /hint %}}

## Usage

Generate examples in place for the whole document:

```bash
go-asyncapi doc gen-examples asyncapi.yaml
```

Generate examples only for all messages in components section (and their nested schemas):

```bash
go-asyncapi doc gen-examples 'asyncapi.yaml#/components/messages'
```

Generate examples only for messages in components section excluding nested jsonschema objects:

```bash
go-asyncapi doc gen-examples --only-messages 'asyncapi.yaml#/components/messages'
```

Write the result to a separate file, leaving the original untouched:

```bash
go-asyncapi doc gen-examples asyncapi.yaml -o asyncapi.examples.yaml
```

Generate three examples for every schema in document, appending to the ones already present:

```bash
go-asyncapi doc gen-examples asyncapi.yaml --only-schemas --count 3 --append
```

## Data format

The `format` field of a JSON Schema drives how a string value is generated. For the date/time formats you can
override the layout with a [Go time format](https://pkg.go.dev/time#pkg-constants):

- `--date-format` — used for `date` fields (default `2006-01-02`).
- `--time-format` — used for `time` fields (default `15:04:05`).
- `--datetime-format` — used for `date-time` fields (default `2006-01-02T15:04:05Z07:00` - [RFC3339](https://datatracker.ietf.org/doc/rfc3339/)).

```bash
go-asyncapi doc gen-examples asyncapi.yaml --datetime-format '2006-01-02 15:04:05'
```
