---
title: "Websocket"
description: "Websocket is supported by go-asyncapi"
---

# Websocket

{{< hint info >}}

{{< figure src="/images/websocket.svg" alt="Websocket" class="text-initial" >}}

**[Websocket](https://en.wikipedia.org/wiki/WebSocket)** is a communication protocol that provides full-duplex 
communication channels over a single TCP connection. It is widely used in web applications to enable 
real-time communication between clients and servers. Websocket allows for bidirectional communication, 
low latency, and efficient data transfer, making it an ideal choice for interactive applications, online games, 
chat applications, and other scenarios that require real-time updates.

Websocket is supported by most modern web browsers and servers, and it is commonly used in conjunction with
HTTP to establish a persistent connection between clients and servers.

{{< /hint >}}

| Feature      | Protocol specifics |
|--------------|--------------------|
| Protocol key | `websocket`        |
| Channel      | Connection         |
| Server       | Websocket server   |
| Envelope     | Websocket Message  |

Protocol bindings are described in https://github.com/asyncapi/bindings/blob/master/websockets/README.md
