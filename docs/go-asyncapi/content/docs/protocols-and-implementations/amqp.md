---
title: "AMQP"
description: "AMQP or Advanced Message Queuing Protocol is supported by go-asyncapi"
---

# AMQP

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
