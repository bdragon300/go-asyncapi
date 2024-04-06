---
title: "Apache Kafka"
# bookFlatSection: false
# bookToc: true
# bookHidden: false
# bookCollapseSection: false
# bookComments: false
# bookSearchExclude: false
---

# Apache Kafka

<link rel="stylesheet" href="/css/text.css">

{{< hint info >}}

{{< figure src="/images/kafka.svg" alt="Apache Kafka" class="text-initial" >}}

**[Apache Kafka](https://kafka.apache.org/)** is a distributed streaming platform that allows for the efficient 
processing of large volumes of real-time data. It enables applications to publish, subscribe to, store, and 
process streams of records in a fault-tolerant and scalable manner. Kafka is widely used for building real-time 
data pipelines, streaming analytics, and event-driven architectures.

{{< /hint >}}

| Feature      | Protocol specifics |
|--------------|--------------------|
| Protocol key | `kafka`            |
| Channel      | Topic              |
| Server       | Kafka server       |
| Envelope     | Kafka Message      |

Protocol bindings are described in https://github.com/asyncapi/bindings/blob/master/kafka/README.md
