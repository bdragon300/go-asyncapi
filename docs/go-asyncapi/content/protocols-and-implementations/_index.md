+++
title = 'Protocols and implementations'
weight = 200
bookToC = true
description = 'The list of protocols and their implementations supported by go-asyncapi'
+++

# Protocols and implementations

## AMQP

{{< hint info >}}

{{< figure src="/images/amqp.svg" alt="AMQP" class="text-initial" >}}

**[AMQP](https://www.amqp.org/)**, or Advanced Message Queuing Protocol, is a standardized messaging protocol
designed for exchanging messages between applications or services. It enables reliable communication across
different platforms and languages, facilitating the building of distributed systems. AMQP defines a set of
rules and formats for message queuing, ensuring interoperability and scalability in complex network environments.

Some of the most commonly used AMQP products include RabbitMQ, Apache Qpid, ActiveMQ and others.

{{< /hint >}}

| Feature       | Protocol specifics                     |
|---------------|----------------------------------------|
| Protocol name | `amqp`                                 |
| Channel       | Exchange (outgoing) / Queue (incoming) |
| Server        | AMQP broker                            |
| Envelope      | AMQP Message                           |

Protocol bindings are described in https://github.com/asyncapi/bindings/tree/v3.0.0/amqp/README.md

## Apache Kafka

{{< hint info >}}

{{< figure src="/images/kafka.svg" alt="Apache Kafka" class="text-initial" >}}

**[Apache Kafka](https://kafka.apache.org/)** is a distributed streaming platform
that allows efficient process large volumes of real-time data.
It enables applications to publish, subscribe to, store, and process streams of records in a fault-tolerant and
scalable manner.
Kafka is widely used for building real-time data pipelines, streaming analytics, and event-driven architectures.

{{< /hint >}}

| Feature       | Protocol specifics |
|---------------|--------------------|
| Protocol name | `kafka`            |
| Channel       | Topic              |
| Server        | Kafka broker       |
| Envelope      | Kafka Message      |

Protocol bindings are described in https://github.com/asyncapi/bindings/tree/v3.0.0/kafka/README.md

## HTTP

{{< hint info >}}

{{< figure src="/images/http.svg" alt="HTTP" class="text-initial" >}}

**[HTTP](https://developer.mozilla.org/en-US/docs/Web/HTTP)**, or Hypertext Transfer Protocol, is an
application-layer protocol for transmitting hypermedia documents, such as HTML. It is the foundation of data
communication on the World Wide Web and is used to exchange information between clients and servers.

HTTP follows a request-response model, where clients send requests to servers and servers respond with messages. It
supports various methods, such as GET, POST, PUT, DELETE, and others, to perform different actions on resources.

Some of the most popular HTTP servers include Apache HTTP Server, Nginx, Microsoft IIS, and others.

{{< /hint >}}

| Feature       | Protocol specifics    |
|---------------|-----------------------|
| Protocol name | `http`                |
| Channel       | HTTP route            |
| Server        | HTTP server           |
| Envelope      | HTTP request/response |

Protocol bindings are described in https://github.com/asyncapi/bindings/tree/v3.0.0/http/README.md

## IP RAW sockets

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

| Feature       | Protocol specifics |
|---------------|--------------------|
| Protocol name | `ip`               |
| Channel       | IP peers pair info |
| Server        | IP peer            |
| Envelope      | IP packet          |

### Bindings

#### Server Bindings

Does not support any server bindings.

#### Channel Bindings

Does not support any channel bindings.

#### Operation Bindings

Does not support any operation bindings.

#### Message Bindings

Does not support any message bindings.

## MQTT

{{< hint info >}}

{{< figure src="/images/mqtt.svg" alt="MQTT" class="text-initial" >}}

**[MQTT](https://mqtt.org/)**, or Message Queuing Telemetry Transport, is a lightweight messaging protocol
designed for constrained devices and low-bandwidth, high-latency, or unreliable networks. It is widely used in
the Internet of Things (IoT) and mobile applications to enable communication between devices and servers.

MQTT follows a publish-subscribe model, where clients publish messages to topics and subscribe to receive messages
from topics. It is known for its simplicity, efficiency, and reliability, making it an ideal choice for IoT
applications and other scenarios with limited resources.

Some of the most popular MQTT brokers include Mosquitto, HiveMQ, EMQ X and others.

{{< /hint >}}

| Feature       | Protocol specifics |
|---------------|--------------------|
| Protocol name | `mqtt`             |
| Channel       | Topic              |
| Server        | MQTT broker        |
| Envelope      | MQTT Message       |

Protocol bindings are described in https://github.com/asyncapi/bindings/tree/v3.0.0/mqtt/README.md

## Redis

{{< hint info >}}

{{< figure src="/images/redis.svg" alt="Redis" class="text-initial" >}}

**[Redis](https://redis.io/)** is an open-source, in-memory data structure store that can be used as a database,
cache, and message broker. It supports various data structures such as strings, hashes, lists, sets, sorted sets,
bitmaps, hyperloglogs, geospatial indexes, and streams. Redis is known for its high performance, scalability,
and versatility, making it a popular choice for real-time applications, caching, and message queuing.

{{< /hint >}}

| Feature       | Protocol specifics |
|---------------|--------------------|
| Protocol name | `redis`            |
| Channel       | Server connection  |
| Server        | Redis server       |
| Envelope      | Redis Message      |

Protocol bindings are described in https://github.com/asyncapi/bindings/tree/v3.0.0/redis/README.md

## TCP

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

| Feature       | Protocol specifics |
|---------------|--------------------|
| Protocol name | `tcp`              |
| Channel       | Connection         |
| Server        | TCP peer           |
| Envelope      | TCP packet         |

### Bindings

#### Server bindings

Does not support any server bindings.

#### Channel bindings

Does not support any channel bindings.

#### Operation bindings

Does not support any operation bindings.

#### Message bindings

Does not support any message bindings

## UDP

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

| Feature       | Protocol specifics  |
|---------------|---------------------|
| Protocol name | `udp`               |
| Channel       | UDP peers pair info |
| Server        | UDP peer            |
| Envelope      | UDP datagram        |

### Bindings

#### Server bindings

Does not support any server bindings.

#### Channel bindings

Does not support any channel bindings.

#### Operation bindings

Does not support any operation bindings.

#### Message bindings

Does not support any message bindings.

## WebSocket

{{< hint info >}}

{{< figure src="/images/websocket.svg" alt="WebSocket" class="text-initial" >}}

**[WebSocket](https://en.wikipedia.org/wiki/WebSocket)** is a communication protocol that provides full-duplex
communication channels over a single TCP connection. It is widely used in web applications to enable
real-time communication between clients and servers. WebSocket allows for bidirectional communication,
low latency, and efficient data transfer, making it an ideal choice for interactive applications, online games,
chat applications, and other scenarios that require real-time updates.

WebSocket is supported by most modern web browsers and servers, and it is commonly used in conjunction with
HTTP to establish a persistent connection between clients and servers.

{{< /hint >}}

| Feature       | Protocol specifics |
|---------------|--------------------|
| Protocol name | `ws`               |
| Channel       | Connection         |
| Server        | Websocket server   |
| Envelope      | Websocket Message  |

Protocol bindings are described in https://github.com/asyncapi/bindings/tree/v3.0.0/websockets/README.md
