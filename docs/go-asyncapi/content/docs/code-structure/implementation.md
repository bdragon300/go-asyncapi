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

Implementation is a concrete library to talk with servers and keep the channels working. The tool has the set of
predefined implementations for all supported protocols, that can be chosen by cli flags, but you can also implement
your own. Depending on what library your project uses, you can choose the implementation that suits you best.

Every implementation are provided the following:

* **Producer** and **Consumer** -- "connection" to a server, that is used to open channels. Can be a stub and just keep
  the connection information if a protocol/library doesn't support a separate connection to the server (e.g. HTTP).
* **Publisher** and **Subscriber** -- "connection" to a channel (topic, queue, etc.), that is used to send and
  receive **Protocol envelopes**.
* **Incoming envelope**, **Outgoing envelope** -- library-specific implementation of **Protocol envelope**, that
  usually contains a message payload, headers, protocol bindings, etc.

