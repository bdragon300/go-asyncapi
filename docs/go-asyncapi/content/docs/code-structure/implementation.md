---
title: "Implementations"
weight: 1
# bookFlatSection: false
# bookToc: true
# bookHidden: false
# bookCollapseSection: false
# bookComments: false
# bookSearchExclude: false
---

## Implementation

Implementation is a concrete library to work with a particular protocol. `go-asyncapi` has the set of
predefined implementations for all supported protocols, that can be chosen by cli flags, but you can also implement
your own.

It's possible to generate an implementation only, without other code. This is useful when you need just a sample
implementation for a protocol, e.g. to make your own based on it. 

The generated implementation code is put by default to `impl` package in target directory.

Every implementation provide the following:

* **Producer** and **Consumer** -- "connection" to a server, that is used to open channels. Can be a stub and just keep
  the connection information if a protocol/library doesn't support a separate connection to the server (e.g. HTTP).
* **Publisher** and **Subscriber** -- "connection" to a channel (topic, queue, etc.), that is used to send and
  receive **Protocol envelopes**.
* **Incoming envelope**, **Outgoing envelope** -- library-specific implementation of 
  [Envelope]({{< relref "/docs/code-structure/message#envelope" >}}), that wraps a message to be able to send 
  or receive it over a particular protocol channel.

{{< details "Usage example" >}}

Take, for example, the Kafka producer provided by the implementation [franz-go](https://github.com/twmb/franz-go).

This is how to use the implementation alone:

```go
import (
    implKafka "myproject/impl/kafka"
)

//...

// Open a connection to the Kafka server
producer, err := implKafka.NewProducer("localhost:9092", nil, nil)
if err != nil {
    log.Fatalf("Failed to create Kafka producer: %v", err)
}

// Connect to a topic
publisher, err := producer.Publisher(context.Background(), "my-topic", nil)
if err != nil {
    log.Fatalf("Failed to create Kafka publisher: %v", err)
}
defer publisher.Close()

// Create an envelope
envelope := implKafka.NewEnvelopeOut()
_, err := envelope.Write([]byte("Hello, Kafka!"))
if err != nil {
    log.Fatalf("Failed to write to Kafka envelope: %v", err)
}

```
{{< /details >}}
