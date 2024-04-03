---
title: "Your own implementation"
weight: 1
# bookFlatSection: false
# bookToc: true
# bookHidden: false
# bookCollapseSection: false
# bookComments: false
# bookSearchExclude: false
---

# Writing your own protocol implementation

Most protocols require two kind of connections for operating: one is a network connection to a server and another
is a channel inside this connection. So, the *Producer*/*Consumer* represent the former, and
the *Publisher*/*Subscriber* represent the latter.

## Server = *Producer+Consumer*

So, the *Producer/Consumer* is something that:

* represents a connection to a server
* responsible for creating *Publishers* and *Subscribers*
* accepts a server URL

The server URL typically contains the server address and port, and may contain other parameters. Also, the server URL
may accept *Server parameters*.

*Producer* typically is a connection intended for publishing data and should implement the following interface:

```go
type Producer interface {
    Publisher(ctx context.Context, channelName string, bindings *ChannelBindings) (Publisher, error)
}
```

*Consumer* typically is a connection intended for consuming data and should implement the following interface:

```go
type Consumer interface {
    Subscriber(ctx context.Context, channelName string, bindings *ChannelBindings) (Subscriber, error)
}
```

`ChannelBindings` type is protocol-specific.

Many libraries treat the *Producer* and *Consumer* as a single entity, so it can be the same object. Others 
use different objects for publishing and consuming, especially for brokerless protocols (HTTP, Websocket, etc.).
This aspect fully depends on a particular library.

## Channel = *Publisher+Subscriber*

*Publisher/Subscriber* typically is:

* represents a channel inside a server connection
* responsible for sending and receiving messages
* accepts a channel name

According to the AsyncAPI specification, the channel name may mean different things for different protocols: a topic
for Apache Kafka, a path for HTTP, a queue/exchange for AMQP, etc. And also, the channel name may accept 
*Channel parameters*.

*Publisher* typically is a channel intended for publishing data and should implement the following interface:

```go
type Publisher interface {
    Send(ctx context.Context, envelopes ...EnvelopeWriter) error
    Close() error
}
```

*Subscriber* typically is a channel intended for consuming data and should implement the following interface:

```go
type Subscriber interface {
    Receive(ctx context.Context, cb func(envelope EnvelopeReader)) error
    Close() error
}
```

`EnvelopeWriter` and `EnvelopeReader` types are protocol-specific interfaces (see below).

Same as before, many libraries treat the *Publisher* and *Subscriber* as a single entity, so it can be the same object.
Others use different objects for sending and receiving messages.

## Message+Protocol = *Envelope*

A message can't just being sent to a concrete server as is, it must contain protocol-specific information. 
In the same time a message is supposed by AsyncAPI authors to be protocol-agnostic.

So, the solution is to wrap protocol-agnostic message by a library-specific *Envelope*. And operate with *Envelopes*.

Many libraries use the same type for incoming and outgoing messages, but some of them use different types.
So we have two basic interfaces for *Envelopes*, one for outgoing data and another for incoming data:

```go
type EnvelopeWriter interface {
    Write(p []byte) (n int, err error)
    ResetPayload()
    SetHeaders(headers Headers)
    SetContentType(contentType string)
}

type EnvelopeReader interface {
    Read(p []byte) (int, error)
    Headers() Headers
}
```

As soon as the *Envelope* is protocol-specific, it can have more methods. For example, `EnvelopeWriter` for 
Apache Kafka has also `SetTopic(topic string)` that is called during preparation an envelope for sending, 
because every single outgoing Kafka message must be assigned to a topic, despite that the topic actually is a 
part of channel information.

## Comments

However, not all protocols obey the approach described above by their design.

### Brokerless (peer-to-peer) protocols

*Websocket* is a brokerless protocol, that implies only one connection directly to the server without any channels 
inside. How can it look like in terms of AsyncAPI?

Everything is simple for *Producer* -- it does nothing. *Publisher* is an opened connection.

But for the *Consumer* the situation is slight complicated: "channel" here is incoming connection that has came 
to a particular HTTP path. So, one of the solutions is make *Consumer* the `http.ServeMux` that intended to 
be passed to your HTTP server object. And then, as soon as a new connection has came, we create a new `Subscriber`.

This situation is typical for HTTP, Websocket, TCP, and other brokerless protocols. We can't just open a channel
and just wait for data. Instead, we must wait for a new channel will be opened as soon as a new connection has came.

### UDP and IP raw sockets

{{< hint warning >}}
The IP and UDP protocols are not described in the AsyncAPI specification. But the specification permits the use of
custom protocols, they can be used in many applications, so it is supported by `go-asyncapi`.
{{< /hint >}}

These protocols do not imply connections at all. So, this case is similar to the previous one, except that *Publisher*
and *Subscriber* don't keep connection opened. The *Consumer* listens to the particular IP/port (UDP) or 
just IP (IP raw sockets).
