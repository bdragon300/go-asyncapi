+++
title = 'Overview'
weight = 10
+++

# Overview

## Initial requirements

The `go-asyncapi` tool has been designed to produce the code, on one hand, that implements all AsyncAPI features, and on 
the other hand, that is modular, extensible and easy to use. 

One of the requirements is that the user should have the choice, if he wants to use the whole generated code or only a 
part of it, or even to replace some parts with his own version.

Many complex architectures may be described in dozens of AsyncAPI documents referenced to each other. So, another 
requirement is to be able to track and handle these dependencies.

Finally, to help the user to mock up or make prototypes or write the applications, it's better to provide minimal 
implementations based on popular libraries for the most popular protocols. And they should be pluggable as well.

In other words, *batteries included, but removable* :)

## Structure

The code structure, reflecting all AsyncAPI features supported by `go-asyncapi`, is shown in figure below.

{{< figure src="../objects.svg" alt="Objects structure diagram" >}}

Let's take a look what each of these objects is.

**[Server]({{< relref "/docs/code-structure/server" >}})** is generated from `server` AsyncAPI entity. 
The main purpose of server is to keep the information to easily open a **Channel** to this server. For this, 
it keeps inside the concrete **Protocol consumer** and **Protocol producer** objects. **Server bindings** contain 
the protocol-specific settings for this server.

**[Channel]({{< relref "/docs/code-structure/channel" >}})** is generated as a separate struct for every 
protocol of the servers bound with this channel, a **Protocol channel** (e.g. `MyChannelKafka`, `MyChannelAMQP`, etc.).
The **Channel** also uses **Channel Parameters** that are used to substitute values to the channel name. 
**Channel bindings**, **Operation bindings** contain the protocol-specific settings for a channel and its operations. 
The `operation` AsyncAPI entity is a part of channel code, so it's not shown.

**[Message]({{< relref "/docs/code-structure/message" >}})** is generated from `message` AsyncAPI definition. The message 
is wrapped by **Protocol envelope** to be able to be sent or received over a particular protocol channel. Channel 
can have different messages for publishing and subscribing, so the **Message** is split up into incoming and 
outgoing part. Due to differences between incoming and outgoing envelope types in many implementations, the 
**Protocol envelope** is also split up into incoming and outgoing parts. **Message bindings** contain the 
protocol-specific settings for this message 

**[Model]({{< relref "/docs/code-structure/model" >}})** is any data structure crafted from 
[jsonschema](https://json-schema.org/) defined in YAML format in `components.schemas` AsyncAPI section. 
Can be referred by any other entity.

**[Implementation]({{< relref "/docs/code-structure/implementation" >}})** is a concrete library to work with a 
particular protocol. 
See TODO
