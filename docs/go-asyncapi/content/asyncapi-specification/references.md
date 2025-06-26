---
title: "References"
weight: 310
description: "External and internal references resolution mechanism description" 
---

# References

## Introduction

[References](https://www.asyncapi.com/docs/concepts/asyncapi-document/reusable-parts) are a powerful feature of the 
AsyncAPI specification that allows you to reuse objects defined in the same or another document. 
This helps to avoid duplication and makes the document more readable and maintainable.

The reference is an object that is denoted by field `$ref` and contains a 
[JSON Reference](https://tools.ietf.org/html/draft-pbryan-zyp-json-ref-03) to the object that should be used instead of the reference.
The value of the `$ref` field can point to an object in the same document, another document by file path, or a remote document by URL.

{{< hint info >}}
Because the reference is URI, the symbols in that are non-alphanumeric and not `-`, `.`, `_`, `~`, must be 
percent-encoded as described in [RFC 3986](https://tools.ietf.org/html/rfc3986#section-2.1)
([encoding table](https://www.w3schools.com/tags/ref_urlencode.ASP)).

For example, the reference that points to a channel with name `foo/bar baz` should be written as
`$ref: "#/components/channels/foo%2Fbar%20baz`.
{{< /hint >}}

{{< hint warning >}}
Remote references initially are forbidden by security reasons, use `--allow-remote-refs` cli flag to allow it.
{{< /hint >}}

For more information about references, see the 
[AsyncAPI documentation](https://www.asyncapi.com/docs/concepts/asyncapi-document/reusable-parts).

## Reference locator

The `go-asyncapi` tool supports references in the AsyncAPI documents and resolves them automatically.

The first step of reference handling is to locate and load the referenced document, which is done by the 
component called **reference locator**.

`go-asyncapi` has a built-in reference locator. If a reference points to the local document (internal reference), e.g. 
`$ref: "#/components/schemas/MySchema"` (which is true in most cases), it does nothing. 

Otherwise, there are two options described below.

### Reference to the file in the filesystem

In this case, the locator just reads the file from the local filesystem. The file path can be relative or absolute:

* Relative path, e.g. `$ref: "foo/bar.yaml#/components/schemas/MySchema"`
* Absolute path, e.g. `$ref: "/path/to/foo/bar.yaml#/components/schemas/MySchema"`
* File URL, e.g. `$ref: "file:///path/to/foo/bar.yaml#/components/schemas/MySchema"`

{{< hint info >}}
The files, which are referenced by the `$ref` field, are resolved relatively to the search directory (current working
directory by default), not to the file where the reference is placed.
Use `--locator-search-dir` cli flag to change this directory.

For example, the reference `foo.yaml#/bar/baz` inside the `/path/to/spam.yaml` will be resolved as
`/search_dir/foo.yaml#/bar/baz`, not as `/path/to/foo.yaml#/bar/baz`. **The second option is not supported yet.**
{{< /hint >}}

### Reference to the remote document by URL

Only HTTP(S) URLs are supported by the built-in reference locator. For example, 
`$ref: "https://example.com/path/to/file#/components/schemas/MySchema"`. 
No authentication or request customization is supported yet.

## Custom reference locator

You can provide your own custom reference locator by passing it in that will be used instead of the built-in one.
Custom locator is an executable that reads a file path from STDIN, does its job, returns the contents of one file on 
STDOUT and exits with 0 return code on success.
If the command returns a non-zero exit code, `go-asyncapi` will exit immediately.

`go-asyncapi` waits for the command to be finished in default timeout of 30 seconds (which can be configured by the
`--locator-timeout` flag or in config file). If a resolver process is still running after this timeout has passed, 
`go-asyncapi` performs the *graceful shutdown*:

1. `go-asyncapi` sends **SIGTERM** signal to a locator process, awaiting it to be finished
2. After another 3 seconds `go-asyncapi` kills a process if it still doesn't respond

Custom locator is passed by the `--locator-command` CLI flag or can be set in configuration:

```yaml
locator:
  command: my-locator.sh
```

{{% hint info %}}
Command flags may be passed in the same way as in the shell, for example:

```yaml
locator:
  command: "my-locator.sh --foo \"Hello world\" -x"
```
{{% /hint %}}

### Example

Let's write a custom resolver in shell language that:

* read a file by absolute path from the local file system
* read a file by relative path from [apicurio schema registry](https://www.apicur.io/registry/)
* fetches HTTP URLs using [curl](https://curl.se/)

```shell
#!/bin/sh

set -e

APICURIO_API_URL="https://registry.apicur.io/apis/registry/v2/groups/asyncapi/artifacts"

# Read the spec file path from STDIN
read FILE_PATH

# Resolve the path
case "$FILE_PATH" in
    "https://"*|"http://"*)  # Read file by HTTP URL
        curl -s --fail-with-body "$FILE_PATH"
        ;;
    /*)  # Read file by absolute path from local file system
        cat "$FILE_PATH"
        ;;
    *)  # Read file by relative path from the apicurio schema registry
        URL="$APICURIO_API_URL/$FILE_PATH"
        curl -s --fail-with-body "$URL"
        ;;
esac
```

Give it executable permissions:

```shell
chmod +x my-locator.sh
```

Now run the code generation:

```shell
go-asyncapi code --allow-remote-refs --locator-command my-locator.sh asyncapi-document.yaml`
```
