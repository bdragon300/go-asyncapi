---
title: "TCP"
# bookFlatSection: false
# bookToc: true
# bookHidden: false
# bookCollapseSection: false
# bookComments: false
# bookSearchExclude: false
---

# TCP

<link rel="stylesheet" href="/css/text.css">

{{< hint info >}}

{{< figure src="/images/tcpudp.svg" alt="TCP" class="text-initial" >}}

**[TCP](https://en.wikipedia.org/wiki/Transmission_Control_Protocol)** (Transmission Control Protocol) is a core 
protocol of the Internet Protocol Suite, responsible for establishing and maintaining connections between 
devices on a network. It ensures reliable, ordered, and error-checked delivery of data packets over IP networks.

{{< /hint >}}

{{< hint warning >}}

**TCP** is not described in the AsyncAPI specification. But the specification permits the use of custom protocols,
and pure TCP can be used in many applications, so it is supported by `go-asyncapi`.

{{< /hint >}}

| Feature      | Protocol specifics |
|--------------|--------------------|
| Protocol key | `tcp`              |
| Channel      | Connection         |
| Server       | TCP peer           |
| Envelope     | TCP packet         |

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