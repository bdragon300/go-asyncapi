---
title: "validate"
weight: 321
description: "Validating AsyncAPI documents"
---

# Validating AsyncAPI documents

`go-asyncapi doc validate` command validates one or more AsyncAPI documents against the
[AsyncAPI 3.x JSON Schema](https://github.com/asyncapi/spec-json-schemas). YAML and JSON are supported.

By default, the command uses the schema bundled in `go-asyncapi` matching the `asyncapi` version declared in the
document. To validate against the custom schema file, pass file with the `-s` option.

{{% hint tip %}}
Built-in schemas are bundled only for certain AsyncAPI versions, run `go-asyncapi doc validate --list` to see the bundled versions.
{{% /hint %}}

If all documents are valid, the command exits successfully. Otherwise, it prints the validation errors and
exits with a non-zero status code.

## Usage

To validate a document or documents against the matching built-in schema:

```bash
go-asyncapi doc validate my-app.yaml your-app.json
```
