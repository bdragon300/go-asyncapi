---
title: "HTTP"
description: "HTTP is supported by go-asyncapi"
---

# HTTP

{{< hint info >}}

{{< figure src="/images/http.svg" alt="HTTP" class="text-initial" >}}

**[HTTP](https://developer.mozilla.org/en-US/docs/Web/HTTP)**, or Hypertext Transfer Protocol, is an 
application-layer protocol for transmitting hypermedia documents, such as HTML. It is the foundation of data 
communication on the World Wide Web and is used to exchange information between clients and servers.

HTTP follows a request-response model, where clients send requests to servers and servers respond with messages. It
supports various methods, such as GET, POST, PUT, DELETE, and others, to perform different actions on resources.

Some of the most popular HTTP servers include Apache HTTP Server, Nginx, Microsoft IIS, and others.

{{< /hint >}}

| Feature      | Protocol specifics |
|--------------|--------------------|
| Protocol key | `http`             |
| Channel      | HTTP route         |
| Server       | HTTP server        |
| Envelope     | HTTP request       |

Protocol bindings are described in https://github.com/asyncapi/bindings/blob/master/http/README.md
