---
title: "IP RAW sockets"
description: "RAW sockets are supported by go-asyncapi"
---

# IP RAW sockets

{{< hint info >}}

{{< figure src="/images/ip.png" alt="IP RAW sockets" class="text-initial" >}}

**[IP RAW sockets](https://en.wikipedia.org/wiki/Raw_socket)**, provide a way to send 
and receive data at the Internet Protocol (IP) level. They allow you to bypass the standard TCP/IP stack and 
interact directly with the network layer.

RAW sockets are used for low-level network operations, such as network scanning, packet crafting, and protocol
development. They are commonly used by network administrators, security professionals, and developers to perform
advanced network tasks.

{{< /hint >}}

{{< hint warning >}}

**IP RAW sockets** are not described in the AsyncAPI specification. But the specification permits the use of custom protocols,
and raw sockets can be used in some applications, so they are supported by `go-asyncapi`.

{{< /hint >}}

| Feature      | Protocol specifics |
|--------------|--------------------|
| Protocol key | `ip`               |
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
