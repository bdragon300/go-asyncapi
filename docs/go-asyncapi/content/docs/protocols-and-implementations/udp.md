---
title: "UDP"
weight: 1
# bookFlatSection: false
# bookToc: true
# bookHidden: false
# bookCollapseSection: false
# bookComments: false
# bookSearchExclude: false
---

# UDP

<link rel="stylesheet" href="/css/text.css">

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

| Field name     | Type    | Description              |
|----------------|---------|--------------------------|
| `localAddress` | string  | Local IP address to use. |
| `localPort`    | integer | Local port to use.       |

### Operation bindings

Does not support any operation bindings.

### Message bindings

Does not support any message bindings.
