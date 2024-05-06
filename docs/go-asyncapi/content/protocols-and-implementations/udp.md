---
title: "UDP"
description: "Pure UDP or User Datagram Protocol is supported by go-asyncapi"
---

# UDP

{{< hint info >}}

{{< figure src="/images/tcpudp.svg" alt="UDP" class="text-initial" >}}

**[UDP](https://en.wikipedia.org/wiki/User_Datagram_Protocol)** (User Datagram Protocol) is a connectionless protocol that
sends data packets, called datagrams, over the network without establishing a connection. It is faster and more efficient
than TCP but does not guarantee the delivery of packets. It is used in applications where speed is more important than
reliability, such as online games, video streaming, and voice over IP (VoIP).

{{< /hint >}}

{{< hint warning >}}

**UDP** is not described in the AsyncAPI specification. But the specification permits the use of custom protocols,
and pure UDP can be used in many applications, so it is supported by `go-asyncapi`.

{{< /hint >}}

| Feature      | Protocol specifics  |
|--------------|---------------------|
| Protocol key | `udp`               |
| Channel      | UDP peers pair info |
| Server       | UDP peer            |
| Envelope     | UDP datagram        |

## Bindings

### Server bindings

Does not support any server bindings.

### Channel bindings

Does not support any channel bindings.

### Operation bindings

Does not support any operation bindings.

### Message bindings

Does not support any message bindings.
