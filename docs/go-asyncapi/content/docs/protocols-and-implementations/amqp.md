---
title: "AMQP"
# bookFlatSection: false
# bookToc: true
# bookHidden: false
# bookCollapseSection: false
# bookComments: false
# bookSearchExclude: false
---

# AMQP

<link rel="stylesheet" href="/css/text.css">

{{< hint info >}}

{{< figure src="/images/amqp.svg" alt="AMQP" class="text-initial" >}}

**[AMQP](https://www.amqp.org/)**, or Advanced Message Queuing Protocol, is a standardized messaging protocol 
designed for exchanging messages between applications or services. It enables reliable communication across 
different platforms and languages, facilitating the building of distributed systems. AMQP defines a set of 
rules and formats for message queuing, ensuring interoperability and scalability in complex network environments.

Some of the most commonly used AMQP products include RabbitMQ, Apache Qpid, ActiveMQ and others.

{{< /hint >}}

| Feature      | Protocol specifics |
|--------------|--------------------|
| Protocol key | `amqp`             |
| Channel      | Exchange           |
| Channel      | Queue              |
| Server       | AMQP server        |
| Envelope     | AMQP Message       |

Protocol bindings are described in https://github.com/asyncapi/bindings/blob/master/amqp/README.md
