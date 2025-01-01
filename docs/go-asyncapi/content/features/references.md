---
title: "References"
weight: 310
description: "go-asyncapi can resolve references ($ref) to objects in the same document, in local or remote file. Custom resolver is supported for more complex scenarios" 
---

# References

AsyncAPI specification allows using references to the objects in the document.
This is useful to avoid duplication and to make the document more readable. 

References are set by `$ref` field in the object. The value of this field is a JSON Pointer to the object that should be
used instead of the reference.

`go-asyncapi` supports all kinds of references:

* References to the objects in the same document, e.g. `#/components/schemas/MySchema`
* References to the objects in another document by file path, e.g.
  `/path/to/file#/components/schemas/MySchema` or `file:///path/to/file#/components/schemas/MySchema`
* References to the objects in the remote document by URL, e.g.
  `https://example.com/path/to/file#/components/schemas/MySchema`

{{< hint warning >}}
Remote references initially are forbidden by security reasons, use `--allow-remote-refs` cli flag to allow it.
{{< /hint >}}

{{< hint info >}}
[Reference is URI](https://datatracker.ietf.org/doc/html/draft-pbryan-zyp-json-ref-03#section-3). So the symbols in 
`$ref` field that are non-alphanumeric and not `-`, `.`, `_`, `~` must be percent-encoded as described in
[RFC 3986](https://tools.ietf.org/html/rfc3986#section-2.1) 
([encoding table](https://www.w3schools.com/tags/ref_urlencode.ASP)).

For example, the reference that points to a channel with name `foo/bar baz` could be written as 
`{"$ref": "#/components/channels/foo%2Fbar%20baz"}`.

{{< /hint >}}

## File resolver

The reference resolving process relies on the spec file resolver that reads the contents of files where 
`$ref` are pointed to.

`go-asyncapi` has a built-in file resolver. It just reads the local files by path from filesystem or fetches remote
files by URL using the standard Go's HTTP client. 

{{< hint warning >}}
The files, which are referenced by the `$ref` field, are resolved relatively to the search directory (current working
directory by default), not to the file
where the reference is placed. Use `--resolver-search-dir` cli flag to change this directory.

For example, the reference `foo.yaml#/bar/baz` inside the `/path/to/spam.yaml` will be resolved as
`/search_dir/foo.yaml#/bar/baz`, not as `/path/to/foo.yaml#/bar/baz`. **The second option is not supported yet.**
{{< /hint >}}

If you need different behavior, you can use your own custom file resolver.

## Custom file resolver

Custom resolver is just an executable you provide. For every specification file path to be resolved, the `go-asyncapi` 
runs this executable, feeds a file path to STDIN and expects the resolved content on STDOUT. 
The command should return 0 on success. If the command returns a non-zero exit code, `go-asyncapi` will exit immediately
as well.

{{< hint info >}}
`go-asyncapi` waits for the command to be finished in timeout of 30 seconds (which can be configured by the 
`--resolver-timeout` flag). If a resolver process is still running after this timeout has passed, `go-asyncapi`
does the *graceful shutdown*:

1. `go-asyncapi` sends **SIGTERM** signal to a process awaiting it to be finished
2. After another 3 seconds `go-asyncapi` kills a process if it still doesn't respond
{{< /hint >}}

{{< details "Example" >}}
{{< tabs "1" >}}
{{< tab "my-resolver.sh" >}}

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
{{< /tab >}}
{{< tab "Usage" >}}
```shell
go-asyncapi generate pubsub --allow-remote-refs --resolver-command my-resolver.sh asyncapi-spec.yaml
```
{{< /tab >}}
{{< /tabs >}}
{{< /details >}}