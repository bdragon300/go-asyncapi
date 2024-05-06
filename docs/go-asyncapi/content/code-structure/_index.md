+++
title = 'Code structure'
bookCollapseSection = true
weight = 400
description = 'Generated code design overview and explanation'
+++

# Code structure

## Initial requirements

The `go-asyncapi` tool has been designed to produce the code, on one hand, that implements all AsyncAPI features, and on
the other hand, that is modular, extensible and easy to use.

One of the requirements is that the user should have the choice, if he wants to use the whole generated code or only a
part of it, or even to replace some parts with his own version.

Many complex architectures may be described in dozens of AsyncAPI documents referenced to each other. So, another
requirement is to be able to track and handle these dependencies.

Finally, to help the user to make prototypes or write the applications, it's better to provide minimal
implementations based on popular libraries for the most popular protocols. And they should be pluggable as well.

## Overview

On the figure below it's shown what the code looks like from high level.

{{< figure src="images/code-overview.svg" alt="Code structure overview" >}}

## Structure

If we'd involve <u>all AsyncAPI features</u> supported by `go-asyncapi`, the code structure would 
be as follows:

{{< figure src="images/code-types.svg" alt="Types structure diagram" >}}

This figure helps give an overview of the generated code structure. Let's take a look at what these objects are.

**[Server]({{< relref "/code-structure/server" >}})** is generated from `server` AsyncAPI entity.
The main purpose of server is to keep the information to easily open a **Channel** to this server. For this,
it keeps inside the concrete **Protocol consumer** and **Protocol producer** objects. **Server bindings** contain
the protocol-specific settings for this server.

**[Channel]({{< relref "/code-structure/channel" >}})** is generated as a separate struct for every
protocol of the servers bound with this channel, a **Protocol channel** (e.g. `MyChannelKafka`, `MyChannelAMQP`, etc.).
The **Channel** also uses **Channel Parameters** that are used to substitute values to the channel name.
**Channel bindings**, **Operation bindings** contain the protocol-specific settings for a channel and its operations.
The `operation` AsyncAPI entity is a part of channel code, so it's not shown.

**[Message]({{< relref "/code-structure/message" >}})** is generated from `message` AsyncAPI definition. The message
is wrapped by **Protocol envelope** to be able to be sent or received over a particular protocol channel. Channel
can have different messages for publishing and subscribing, so the **Message** is split up into incoming and
outgoing part. Due to differences between incoming and outgoing envelope types in many implementations, the
**Protocol envelope** is also split up into incoming and outgoing parts. **Message bindings** contain the
protocol-specific settings for this message

**[Model]({{< relref "/code-structure/model" >}})** is any data structure crafted from
[jsonschema](https://json-schema.org/) defined in `components.schemas` AsyncAPI section.
Can be referred by any other entity.

**[Implementation]({{< relref "/code-structure/implementation" >}})** is a concrete library to work with a
particular protocol. See [Protocols and implementations]({{< relref "/protocols-and-implementations" >}}) section
for a full list.

**Encoding** package contains the code used to marshal and unmarshal the messages based on their 
content type. All encoders/decoders of [supported content types]({{< relref "/features#content-types" >}}) 
mentioned in the AsyncAPI document get here. You can add your own encoders/decoders and replace the existing ones.