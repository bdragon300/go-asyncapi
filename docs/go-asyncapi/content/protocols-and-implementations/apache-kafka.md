---
title: "Apache Kafka"
description: "Apache Kafka is supported by go-asyncapi"
---

# Apache Kafka

{{< hint info >}}

{{< figure src="/images/kafka.svg" alt="Apache Kafka" class="text-initial" >}}

**[Apache Kafka](https://kafka.apache.org/)** is a distributed streaming platform
that allows efficient process large volumes of real-time data.
It enables applications to publish, subscribe to, store, and process streams of records in a fault-tolerant and 
scalable manner. 
Kafka is widely used for building real-time data pipelines, streaming analytics, and event-driven architectures.

{{< /hint >}}

| Feature      | Protocol specifics |
|--------------|--------------------|
| Protocol key | `kafka`            |
| Channel      | Topic              |
| Server       | Kafka server       |
| Envelope     | Kafka Message      |

Protocol bindings are described in https://github.com/asyncapi/bindings/blob/master/kafka/README.md
