---
title: "References"
weight: 310
# bookFlatSection: false
# bookToc: true
# bookHidden: false
# bookCollapseSection: false
# bookComments: false
# bookSearchExclude: false
---

# References

AsyncAPI specification allows to use references to the objects in the document. This is useful to avoid duplication and
to make the document more readable. 

References are set by `$ref` field in the object. The value of this field is a JSON Pointer to the object that should be
used instead of the reference.

`go-asyncapi` supports all kinds of references:

* References to the objects in the same document, e.g. `#/components/schemas/MySchema`
* References to the objects in another document by file path, e.g.
  `/path/to/file#/components/schemas/MySchema`
* References to the objects in the remote document by URL, e.g.
  `https://example.com/path/to/file#/components/schemas/MySchema`

Remote references are forbidden by default by security reasons, use `--remote-refs` cli flag to allow it.

## Spec resolver

The reference resolving process relies on the spec resolver that reads the contents of files mentioned in `$ref` values.

`go-asyncapi` has a built-in spec resolver. It just reads the local files by path or fetches remote files by URL 
using the standard Go's HTTP client. 

If you need different behavior, you can use your own custom spec resolver.

## Custom spec resolver

Custom resolver is just an executable. For every specification file path to be resolved, the `go-asyncapi` 
runs this executable, feeds a path to STDIN and expects the resolved content on STDOUT. 
The command should return 0 (within timeout) if the reference is resolved successfully.

{{< details "Example" >}}
{{< tabs "1" >}}
{{< tab "my-resolver.sh" >}}

Let's write a custom resolver in shell language that:

* reads a file by absolute path from the local file system
* reads a file by relative path from [apicurio schema registry](https://www.apicur.io/registry/)
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
go-asyncapi generate pubsub --remote-refs --spec-resolver-command my-resolver.sh asyncapi-spec.yaml
```
{{< /tab >}}
{{< /tabs >}}
{{< /details >}}