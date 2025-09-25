---
title: "Visualization"
weight: 650
bookToC: true
description: "Visualizing AsyncAPI documents in diagrams"
---

# Visualizing AsyncAPI documents

`go-asyncapi` can generate diagrams from AsyncAPI documents using [D2](https://d2lang.com/) -- a modern diagram scripting 
language. The resulting diagrams visualize the structure of the AsyncAPI document, including servers, channels,
operations, messages and their relationships. 

The diagram is generated internally as D2 script. By default, it's additionally converted to SVG format, which allows 
to embed it in web pages, documentation, or share as a standalone file.

The result can also be customized in d2 script templates. See 
[templating guide]({{<relref "/templating-guide/overview">}}) for more details.

## Usage

To generate a diagram, use the `diagram` subcommand:

```bash
$ go-asyncapi diagram my-app.yaml
```

Default result (SVG) will look like this:

{{< figure src="images/my-app-default.svg" alt="my-app.svg" >}}

{{% hint tip %}}
By default, the command generates a file with the same name as the input file, but with `.svg` extension. 
You can change the output file name with the `--output` option. To generate the diagram in D2 format, use the `--format d2`.
{{% /hint %}}


## Customization

The diagram generation process has several built-in customization options.

The default draws operations as "endpoints" and both channels and servers as central nodes. 
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
