---
title: "diagram"
weight: 340
bookToC: true
description: "Generating the diagrams"
---

# Visualizing AsyncAPI documents in diagrams

`diagram` command generate diagrams of AsyncAPI entities, that visualize the structure
of the AsyncAPI document, including servers, channels, operations, messages and their relationships.

Under the hood `diagram` command uses the [D2](https://d2lang.com/) diagram scripting language. The diagram is 
generated internally as D2 script and converted to SVG if requested.

The result is an SVG file or D2 script file placed in the current working directory. By default, the output file name is
the same as the input file name, but with `.svg` extension.

By default, the diagram shows entities from a passed AsyncAPI document and all documents that were fetched via `$ref` references.
The output file is named after the passed document with `.svg` or `.d2` extension.
It is possible to split the result into multiple diagrams, change the diagram layout and appearance using various options.

{{% hint tip %}}
The result can also be customized in d2 script templates. See 
[templating guide]({{<relref "/templating-guide/overview">}}) for more details.
{{% /hint %}}

## Usage

{{% hint tip %}}
All command options are duplicated in [configuration]({{<relref "/configuration">}}) file. So, you can set them there as well.
{{% /hint %}}

To generate a diagram, use the `diagram` subcommand:

```bash
$ go-asyncapi diagram my-app.yaml
```

Default result (SVG) will look like this:

{{< figure src="images/my-app-default.svg" alt="my-app.svg" >}}

To split the result into multiple diagrams, one per each input document (the passed one and referenced ones), 
use the `--multiple-files` option:

```bash
$ go-asyncapi diagram my-app.yaml --multiple-files
```

## Customization

By default, the diagram shows operations as "endpoints" and both channels and servers as central nodes. 
Arrows between operations and channels indicate the message name and its flow direction.

{{% hint info %}}
All command-line options mentioned below are also can be set in configuration file, see 
[configuration reference]({{< relref "/configuration" >}})
{{% /hint %}}

### Channel-centric and server-centric views

The option `--channels-centric` makes channels as central nodes, servers are omitted, like shown below:

{{< figure src="images/my-app-channels.svg" alt="my-app.svg" >}}

Another option `--servers-centric` leaves only servers, hiding channels:

{{< figure src="images/my-app-servers.svg" alt="my-app.svg" >}}

### Document borders

By default, the diagram does not show which entities belong to which files. To enable this, use the `--document-borders` option.

{{< figure src="images/my-app-doc-borders.svg" alt="my-app.svg" >}}

### D2 engine options

#### Themes

`--d2-theme-id` option selects a light color theme, while `--d2-dark-theme-id` generates the diagram using dark theme. 
If both options are set, the resulting SVG will be dual-theme where the theme switches automatically based on user's 
system preferences ([more info](https://developer.mozilla.org/en-US/docs/Web/CSS/@media/prefers-color-scheme)).

List of available themes can be found in the [D2 documentation](https://d2lang.com/tour/themes/).

{{< figure src="images/my-app-theme1.svg" alt="my-app.svg" >}}

{{< figure src="images/my-app-theme2.svg" alt="my-app.svg" >}}

#### Other options

[D2](https://d2lang.com/) engine supports various options to customize the appearance of the diagram. See `--help` 
output or [configuration reference]({{< relref "/configuration" >}}) for the full list.

For example, the `--d2-sketch` option makes the diagram look hand-drawn:

{{< figure src="images/my-app-sketch.svg" alt="my-app5.svg" >}}

{{% hint tip %}}
D2 engine sometimes may produce diagrams that looks ugly or that are hard to read due to overlapping nodes or edges.
To get better results, you may try to play with various `--d2-*` CLI args (or appropriate config options), 
starting with `--d2-engine` and `--d2-direction`.
{{% /hint %}}

### More customization

You can create your own D2 templates to customize the diagram output. See 
[templating guide]({{<relref "/templating-guide/overview">}}) for more details.
