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
component called **reference locator**. `go-asyncapi` has a built-in reference locator. 
Locator is intended to find other files, so it skips the reference points to the same document, e.g. 
`$ref: "#/components/schemas/MySchema"`.

### Built-in locator

*Absolute path* to file in filesystem, e.g. `$ref: "/path/to/foo/bar.yaml#/components/schemas/MySchema"` is handled as-is.

*Relative path*, e.g. `$ref: "foo/bar.yaml#/components/schemas/MySchema"` is resolved relative to the current document location. 
If the root directory is set by the `--locator-root-dir` flag or in config file, the path is resolved relative to this directory.

Path can also be a *file URL*, e.g. `$ref: "file:///path/to/foo/bar.yaml#/components/schemas/MySchema"`, which is 
handled the same way as described above.

*HTTP(S) URL* is fetched by built-in Go HTTP client with default config. No authentication or request customization is supported.

{{% hint warning %}}
By default, URL references are forbidden for security reasons, use `--allow-remote-refs` CLI flag or set 
`allowRemoteReferences` option in config file to allow it.
{{% /hint %}}

{{% hint note %}}
Only HTTP(S) remote references are supported by the built-in locator. To support other schemes, you can provide the custom
locator (see below).
{{% /hint %}}

### Custom reference locator

You can provide your own locator by passing a shell command that will be used instead of the built-in one.
`go-asyncapi` runs this command for each handling reference, feed this reference to its STDIN stream and awaits the result from STDOUT.
The command must read reference from STDIN, do its job, return the file contents to STDOUT and exit with 0 return code 
on success. On non-zero exit code, the `go-asyncapi` treats it as locator error and exits immediately with an error.

Once the command launched, `go-asyncapi` waits for it to be finished in default timeout of 30 seconds (configurable by the
`--locator-timeout` flag or in config file). If it is still running after this timeout has passed, 
`go-asyncapi` performs the *graceful shutdown* on command process:

1. `go-asyncapi` sends **SIGTERM** signal to a process, awaiting it to be finished
2. After another 3 seconds `go-asyncapi` kills a process if it still doesn't respond. This outcome is treated as 
   locator error and makes `go-asyncapi` to exit with an error.

Custom locator is passed by the `--locator-command` CLI flag or can be set in configuration:

```yaml
locator:
  command: my-locator.sh
```

Command flags may be passed in the same way as in the shell, for example:

```yaml
locator:
  command: "my-locator.sh --foo \"Hello world\" -x"
```

#### Example

Let's write a custom locator as a shell script that:

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
