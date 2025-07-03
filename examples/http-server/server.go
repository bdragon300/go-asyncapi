package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"http-server/asyncapi/channels"
	http2 "http-server/asyncapi/impl/http"
	"http-server/asyncapi/messages"
	"http-server/asyncapi/schemas"
	"net"
	"net/http"
	"strconv"

	http3 "github.com/bdragon300/go-asyncapi/run/http"
)

func main() {
	port := flag.Int("port", 8080, "Kafka broker port")
	flag.Parse()

	ctx := context.Background()

	// Consumer implements the [http.Handler] interface,
	// so you can use it as a handler like any other HTTP handler. For example:
	//
	//   http.Handle("/prefix", implConsumer)
	//
	// Here we use it as a root handler for the HTTP server.
	implConsumer := http2.NewConsumer(nil)
	httpServer := http.Server{
		Addr:    ":" + strconv.Itoa(*port),
		Handler: implConsumer,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	chAddress := channels.EchoChannelAddress().String()
	implSubscriber, err := implConsumer.Subscriber(ctx, chAddress, nil, nil)
	if err != nil {
		panic(err)
	}
	defer implSubscriber.Close()

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()
	fmt.Printf("Listening on port %d\n", *port)

	// We use the "std" HTTP implementation here ("net/http").
	// So, for request-response model we work with envelopes instead of messages,
	// because "net/http" HTTP server provides no other way to write the response other than the [http.ResponseWriter],
	// passed to the handler as an argument.
	// Envelope encapsulates the [http.Request] and [http.ResponseWriter] passed to handler.
	channel := channels.NewEchoChannelHTTP(nil, implSubscriber)
	err = channel.Subscribe(ctx, func(envelope http3.EnvelopeReader) {
		req := envelope.(*http2.EnvelopeIn)

		// Unseal a message by hand (i.e. read and unmarshal it).
		var reqMsg messages.ServerRequestIn
		if err := channel.UnsealServerRequest(envelope, &reqMsg); err != nil {
			fmt.Println("Error unsealing message: ", err)
			return
		}
		fmt.Printf("Received message: %+v\n", reqMsg.Payload())

		// Craft a response
		respMsg := messages.ServerResponseOut{
			Payload: schemas.EchoResponse{
				Type:    "echoRespose",
				Message: reqMsg.Payload(),
			},
		}

		// Marshal a message with the message's content type marshaller and write it to the response writer.
		req.ResponseWriter.WriteHeader(http.StatusOK)
		if err := respMsg.MarshalHTTP(req.ResponseWriter); err != nil {
			fmt.Print("Error marshalling response: ", err)
			return
		}

		fmt.Printf("Sent response: %+v\n", respMsg.Payload)
	})
	if err != nil {
		panic(fmt.Errorf("error subscribing to channel %s: %w", chAddress, err))
	}
}
