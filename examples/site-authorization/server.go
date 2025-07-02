package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/bdragon300/go-asyncapi/run/kafka"
	"github.com/twmb/franz-go/pkg/kgo"
	"math/rand/v2"
	"net"
	"net/url"
	"os"
	"site-authorization/asyncapi/channels"
	"site-authorization/asyncapi/messages"
	"site-authorization/asyncapi/parameters"
	"site-authorization/asyncapi/schemas"
	"site-authorization/asyncapi/servers"
	"strconv"
)

type ResponseMode int

const (
	ResponseModeEcho = iota
	ResponseModeGlitch
)

func main() {
	host := flag.String("host", "localhost", "Kafka broker host")
	port := flag.Int("port", 9092, "Kafka broker port")
	environment := flag.String("env", "development", "Environment parameter for the channel")
	rMode := flag.String("mode", "echo", "ResponseMode parameter for the channel. Possible values are: echo, glitch")
	flag.Parse()

	var responseMode ResponseMode
	switch *rMode {
	case "echo":
		responseMode = ResponseModeEcho
	case "glitch":
		responseMode = ResponseModeGlitch
	default:
		fmt.Printf("Unknown response mode: %s\n", *rMode)
		os.Exit(1)
	}

	brokerURL := url.URL{Scheme: "kafka", Host: net.JoinHostPort(*host, strconv.Itoa(*port))}
	ctx := context.Background()

	server, err := servers.ConnectTestServerBidi(
		ctx,
		&brokerURL,
		kgo.WithLogger(kgo.BasicLogger(os.Stdout, kgo.LogLevelInfo, nil)),
		kgo.ConsumerGroup("auth-server-"+*environment),
	)
	if err != nil {
		fmt.Printf("Failed to connect to Kafka server: %v\n", err)
		os.Exit(1)
	}
	defer server.Close()

	if err := serverLoop(ctx, server, parameters.EnvironmentName(*environment), responseMode); err != nil {
		fmt.Printf("Failed to send message to Kafka server: %v\n", err)
		os.Exit(1)
	}
}

func serverLoop(ctx context.Context, server channels.AuthChannelServerKafka, env parameters.EnvironmentName, responseMode ResponseMode) error {
	fmt.Println("Awaiting auth requests...")
	channel, err := server.OpenAuthChannelKafka(ctx, channels.AuthChannelParameters{EnvironmentName: env})
	if err != nil {
		return fmt.Errorf("open auth channel: %w", err)
	}

	subCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	err = channel.Subscribe(subCtx, func(e kafka.EnvelopeReader) {
		msg := new(messages.AuthResponseMsgIn)
		if err := channel.UnsealAuthRequestMsg(e, msg); err != nil {
			fmt.Printf("unseal message envelope: %v\n", err)
			return
		}
		if msg.Payload().EventType != "AuthRequest" {
			fmt.Printf("Ignore the message with event type %s\n", msg.Payload().EventType)
			return
		}
		resps, err := handleAuthRequest(msg, responseMode)
		if err != nil {
			fmt.Printf("handle auth request: %v\n", err)
			return
		}
		for _, resp := range resps {
			if err := channel.PublishAuthRequestMsg(subCtx, &resp); err != nil {
				cancel(fmt.Errorf("publish auth response: %w", err))
				return
			}
			fmt.Printf("Published auth response: %+v\n", resp)
		}
	})
	if err != nil {
		return fmt.Errorf("subscribe to auth requests: %w", err)
	}
	return nil
}

func handleAuthRequest(msg messages.AuthRequestMsgSender, responseMode ResponseMode) ([]messages.AuthResponseMsgOut, error) {
	fmt.Printf("Received auth request: %+v\n", msg.Payload())

	var res []messages.AuthResponseMsgOut
	corrID, err := msg.CorrelationID()
	if err != nil {
		return nil, fmt.Errorf("get correlation ID: %w", err)
	}

	switch responseMode {
	case ResponseModeGlitch:
		eventType := "AuthResponse"
		newCorrID := corrID

		switch rand.Int() % 2 {
		case 0:
			eventType = "foo"
		case 1:
			newCorrID = rand.IntN(1000) // Random new correlation ID
		}

		res = append(res, messages.AuthResponseMsgOut{
			Payload: schemas.AuthEvent{
				EventType: eventType,
				ID:        newCorrID,
				AuthResponse: schemas.AuthResponse{
					Message: "Glitch: " + msg.Payload().AuthRequest.Username,
					Success: rand.IntN(2) == 0, // Random success/failure
				},
			},
		})
		fallthrough
	case ResponseModeEcho:
		res = append(res, messages.AuthResponseMsgOut{
			Payload: schemas.AuthEvent{
				EventType: "AuthResponse",
				ID:        corrID,
				AuthResponse: schemas.AuthResponse{
					Message: "Echo: " + msg.Payload().AuthRequest.Username,
					Success: rand.IntN(2) == 0, // Random success/failure
				},
			},
		})
	}

	return res, nil
}