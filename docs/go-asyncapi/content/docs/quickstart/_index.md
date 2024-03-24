---
title: "Quickstart"
weight: 40
# bookFlatSection: false
# bookToc: true
# bookHidden: false
# bookCollapseSection: false
# bookComments: false
# bookSearchExclude: false
---

# Quickstart

This guide will help you to get started with the go-asyncapi tool. In this example we'll create a simple echo Websocket
server that listens for incoming messages and sends them back to the client.

The following AsyncAPI document describes one server and one channel with two operations: `publish` and `subscribe`:

{{< details "test.asyncapi.yaml" >}}
```yaml
asyncapi: '2.6.0'
info:
  title: 'Websocket Echo Service'
  version: 1.0.0

channels:
  wsAPI:
    publish:
      description: Echo request
      message:
        $ref: '#/components/messages/pongMessage'
    subscribe:
      description: Echo response
      message:
        $ref: '#/components/messages/pingMessage'

servers:
  websocketAPI:
    url: ws://api.example.com
    protocol: ws

components:
  messages:
    pingMessage:
      payload:
        $ref: "#/components/schemas/textMessage"
    pongMessage:
      payload:
        type: string

  schemas:
    textMessage:
      type: object
      properties:
        message:
          type: string
```
{{< /details >}}

It's time to generate Go code using the go-asyncapi tool:

```shell
go-asyncapi generate pubsub test.asyncapi.yaml
```

The generated Go code will be placed to **./asyncapi** directory. Let's take a look what we've got:

* **channels/ws_api.go** -- Websocket-specific code for channel `wsAPI`
* **encoding/{encode,decode}.go** -- encoder/decoder for messages. We didn't define the content type in the 
  AsyncAPI document, so the default JSON encoding is used
* **messages/{ping_message,pong_message}.go** -- `pingMessage` and `pongMessage` messages
* **models/text_message.go** -- `textMessage` model that is used in messages
* **servers/websocket_api.go** -- `websocketAPI` server
* **impl/ws/** -- minimal default ready-to-use implementation for work with Websocket protocol

Now, we can write a simple echo server that uses code generated before:

{{< details "main.go" >}}
```go
package main

import (
  "context"
  "fmt"
  "log"
  "net/http"
  "os"
  "os/signal"
  "time"

  implWs "github.com/bdragon300/go-asyncapi/asyncapi/impl/ws"
  "github.com/bdragon300/go-asyncapi/asyncapi/messages"
  "github.com/bdragon300/go-asyncapi/asyncapi/servers"
  runWs "github.com/bdragon300/go-asyncapi/run/ws"
)

func main() {
  cancelCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
  defer cancel()

  // Create an implementation consumer
  consumer, err := implWs.NewConsumer(nil, 5*time.Second)
  if err != nil {
    log.Fatalf("failed to create consumer: %v", err)
  }
  // Pass the implementation consumer to the server object.
  // Producer is nil here, since we only listen for incoming messages and send responses, not initiate
  // new connections (see Websocket protocol docs for more details).
  wsServer := servers.NewWebsocketAPI(nil, consumer)
  // Open a channel
  channel, err := wsServer.OpenWsAPIWebSocket(cancelCtx)
  if err != nil {
    log.Fatalf("failed to open subscribe channel: %v", err)
  }
  defer channel.Close()

  httpServer := http.Server{Addr: ":80", Handler: consumer}
  go func() {
    if err := httpServer.ListenAndServe(); err != nil {
      log.Fatalf("failed to start http server: %v", err)
    }
  }()

  // Blocked call, await for new connections
  err = channel.Subscribe(cancelCtx, func(envelope runWs.EnvelopeReader) {
    // Extract a message from an envelope
    msgIn := messages.NewPingMessageIn()
    if err := channel.ExtractEnvelope(envelope, msgIn); err != nil {
      log.Fatalf("failed to extract envelope: %v", err)
    }
    log.Printf("received message: %v", msgIn.Payload)

    // Make message to send, wrap it into an envelope and send it back
    msgOut := messages.NewPongMessageOut().WithPayload(
      fmt.Sprintf("Hello! Your message was: %v", msgIn.Payload.Message),
    )
    envelopeOut := implWs.NewEnvelopeOut()
    if err := channel.MakeEnvelope(envelopeOut, msgOut); err != nil {
      log.Fatalf("failed to make envelope: %v", err)
    }
    if err := channel.Publish(cancelCtx, envelopeOut); err != nil {
      log.Fatalf("failed to send message: %v", err)
    }
  })
  if err != nil {
    log.Fatalf("failed to subscribe: %v", err)
  }
}
```
{{< /details >}}

Save this code to **main.go** and run it this way:

```shell
go run main.go
```

Finally, let's test our Websocket echo server using the well-known [curl](https://curl.se/) tool. Open a new terminal 
and run this command:

```shell
curl -i -N \
  -H "Connection: Upgrade" \
  -H "Upgrade: websocket" \
  -H "Host: localhost" \
  -H "Origin: http://localhost" \
  http://localhost/wsAPI
```

#TODO: feed input to curl above, test everything above