---
title: "Generate publish only or subscribe only code"
weight: 550
description: "How to generate only publishing or subscribing code"
---

# Generating publish only or subscribe only code

To generate only the publishing or subscribing code, you can use the `--only-pub` and `--only-sub` command-line options
(or appropriate configuration options), respectively.

Example:

```shell
go-asyncapi code --only-pub my_asyncapi.yaml
```

or in the configuration file:

```yaml
code:
  onlyPublish: true
```