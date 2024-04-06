---
title: "Redis"
description: "Redis is supported by go-asyncapi"
---

# Redis

{{< hint info >}}

{{< figure src="/images/redis.svg" alt="Redis" class="text-initial" >}}

**[Redis](https://redis.io/)** is an open-source, in-memory data structure store that can be used as a database, 
cache, and message broker. It supports various data structures such as strings, hashes, lists, sets, sorted sets, 
bitmaps, hyperloglogs, geospatial indexes, and streams. Redis is known for its high performance, scalability, 
and versatility, making it a popular choice for real-time applications, caching, and message queuing.

{{< /hint >}}

| Feature      | Protocol specifics |
|--------------|--------------------|
| Protocol key | `redis`            |
| Channel      | Server connection  |
| Server       | Redis server       |
| Envelope     | Redis Message      |

Protocol bindings are described in https://github.com/asyncapi/bindings/blob/master/redis/README.md
