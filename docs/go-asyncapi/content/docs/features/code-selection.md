---
title: "Code selection"
weight: 320
description: "Objects to generate could be included or excluded from the result by entity kind, name regex or publish/subscribe mark"
---

# Code selection

It's possible to customize what code should be generated or not. For example, you may want to 
generate the code only for the channels, or for a particular model, or for everything except the servers, or only
the subscriber code, etc.

The bunch of cli flags may help to control which code will appear in the result.

## Implementations

### Selection

You can use a built-in implementation for a particular protocol by using `--*-impl` flags. E.g. `--kafka-impl sarama` 
tells the tool to use **Sarama** implementation for Kafka protocol instead of default **franz-go**.

It's possible to generate the implementation only, without other code. This is useful when you need just a sample
implementation for a protocol, e.g. to make your own based on it. Use the `implementation` subcommand: 
`go-asyncapi generate implementation kafka sarama`.

Full list of supported implementations can be found in the 
[Implementations]({{< relref "/docs/features#protocols" >}}) page or by running the 
`go-asyncapi list-implementations` command.

### Exclusion

To refuse from the implementation for a particular protocol, pass the `--kafka-impl none` cli flag. To refuse from 
all implementations pass the `--no-implementations` cli flag, the implementations will not appear at all in this case.

## Entities

### Selection

When you run the `go-asyncapi` tool, you must decide what code you want to generate: publisher, subscriber or both.
The `pub`, `sub` and `pubsub` cli subcommands are used for this purpose.

The group of `--only-*` cli flags selects the appropriate entity types to generate, excluding the others. E.g. 
`--only-channels` generates the code only for the channels. Flags can be combined, e.g. 
`--only-channels --only-schemas` generates the code both for the channels and schemas.

There are special cli flags that select particular AsyncAPI entities. They accept a 
[regular expression](https://en.wikipedia.org/wiki/Regular_expression), and the tool selects only those entities whose names matched this expression. 
These flags also can be combined. Some examples:
 
* `--channels-re '^user''` generates the code only for the channels that names start with **user**. 
* `--models-re '^foobar$'` generates the code only for the models named **foobar**.
* `--servers-re 'prod'` generates the code only for the servers that names contain a substring **prod**.

### Exclusion

`--ignore-*` cli flags, on the contrary, exclude the appropriate entities from the generation, keeping others. E.g. 
`--ignore-channels` generates the code for everything except the channels. Flags can be combined, e.g. 
`--ignore-channels --ignore-schemas` generates the code for everything except the channels and schemas.

`--ignore-*-re` flags tells the tool to exclude the entities whose names match the regular expression. Some examples:

* `--ignore-channels-re '^user''` generates the code for everything except the channels that names start with **user**.
* `--ignore-models-re '^foobar$'` generates the code for everything except the models named **foobar**.
* `--ignore-servers-re 'prod'` generates the code for everything except the servers that names contain a substring **prod**.

It's possible also to exclude the encoding package from the generation, use the `--no-encoding` cli flag for that.
