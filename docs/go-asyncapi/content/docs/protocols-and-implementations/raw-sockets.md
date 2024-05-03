---
title: "Raw sockets"
description: "Raw IP sockets are supported by go-asyncapi"
---

# Raw sockets

{{< hint info >}}

{{< figure src="/images/ip.png" alt="Raw sockets" class="text-initial" >}}

**[Raw sockets](https://en.wikipedia.org/wiki/Raw_socket)**, also known as raw IP sockets, provide a way to send 
and receive data at the Internet Protocol (IP) level. They allow you to bypass the standard TCP/IP stack and 
interact directly with the network layer.

Raw sockets are used for low-level network operations, such as network scanning, packet crafting, and protocol
development. They are commonly used by network administrators, security professionals, and developers to perform
advanced network tasks.

{{< /hint >}}

{{< hint warning >}}

**Raw sockets** are not described in the AsyncAPI specification. But the specification permits the use of custom protocols,
and raw sockets can be used in some applications, so they are supported by `go-asyncapi`.

{{< /hint >}}

| Feature      | Protocol specifics |
|--------------|--------------------|
| Protocol key | `rawsocket`        |
| Channel      | IP peers pair info |
| Server       | IP peer            |
| Envelope     | IP packet          |

## Bindings

### Server Bindings

Does not support any server bindings.

### Channel Bindings

Does not support any channel bindings.

### Operation Bindings

Does not support any operation bindings.

### Message Bindings

Does not support any message bindings.
