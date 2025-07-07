package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	http2 "http-server/asyncapi/impl/http"
	"http-server/asyncapi/messages"
	"http-server/asyncapi/schemas"
	"http-server/asyncapi/servers"
	"net"
	"net/http"
	"strconv"

	http3 "github.com/bdragon300/go-asyncapi/run/http"
)

func main() {
	port := flag.Int("port", 8081, "Kafka broker port")
	flag.Parse()

	ctx := context.Background()
	// For HTTP consumer this function actually doesn't connect anywhere, it's just a protocol-agnostic general function.
	// For HTTP the following gives the same result:
	//
	//   server := http2.NewConsumer(nil)
	//
	server, err := servers.ConnectTestServerConsumer(ctx, nil)
	if err != nil {
		panic(fmt.Errorf("open server: %w", err))
	}
	defer server.Close()

	// Consumer in "std" implementations complies the [http.Handler] interface
	// and can be used as a handler like any other [net/http] handler. For example:
	//
	//   http.Handle("/prefix", consumer)
	//
	// Here we use it as a root handler for the HTTP server.
	consumer := server.Consumer()
	httpServer := http.Server{
		Addr:    ":" + strconv.Itoa(*port),
		Handler: consumer.(http.Handler),
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
	}

	channel, err := server.OpenEchoChannelHTTP(ctx)
	if err != nil {
		panic(fmt.Errorf("open channel: %w", err))
	}
	defer channel.Close()

	go func() {
		if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()
	fmt.Printf("Listening on port %d\n", *port)

	// For request-response model we work with envelopes instead of messages,
	// because [net/http] server provides no other way to write the response other than the [http.ResponseWriter],
	// that being passed to the handler as an argument.
	// Envelope encapsulates the [http.Request] and [http.ResponseWriter] passed to handler.
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
		panic(fmt.Errorf("error subscribing to channel %s: %w", channel.Address(), err))
	}
}
