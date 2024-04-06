---
title: "MQTT"
# bookFlatSection: false
# bookToc: true
# bookHidden: false
# bookCollapseSection: false
# bookComments: false
# bookSearchExclude: false
---

# MQTT

<link rel="stylesheet" href="/css/text.css">

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

| Feature      | Protocol specifics |
|--------------|--------------------|
| Protocol key | `mqtt`             |
| Channel      | Topic              |
| Server       | MQTT server        |
| Envelope     | MQTT Message       |

Protocol bindings are described in https://github.com/asyncapi/bindings/blob/master/mqtt/README.md
