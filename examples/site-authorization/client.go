package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"math/rand/v2"
	"net"
	"net/url"
	"os"
	"site-authorization/asyncapi/channels"
	"site-authorization/asyncapi/messages"
	"site-authorization/asyncapi/operations"
	"site-authorization/asyncapi/parameters"
	"site-authorization/asyncapi/schemas"
	"site-authorization/asyncapi/servers"
	"strconv"

	"github.com/twmb/franz-go/pkg/kgo"
)

func main() {
	host := flag.String("host", "localhost", "Kafka broker host")
	port := flag.Int("port", 9092, "Kafka broker port")
	user := flag.String("user", "testuser", "Username for authentication")
	password := flag.String("password", "testpassword", "Password for authentication")
	environment := flag.String("env", "development", "Environment parameter for the channel")
	flag.Parse()

	brokerURL := url.URL{Scheme: "kafka", Host: net.JoinHostPort(*host, strconv.Itoa(*port))}
	ctx := context.Background()

	// Open a bidirectional connection to the Kafka server
	server, err := servers.ConnectTestServerBidi(
		ctx,
		&brokerURL,
		kgo.WithLogger(kgo.BasicLogger(os.Stdout, kgo.LogLevelInfo, nil)),
		kgo.ConsumerGroup("auth-client-"+*environment),
	)
	if err != nil {
		fmt.Printf("Failed to connect to Kafka server: %v\n", err)
		os.Exit(1)
	}
	defer server.Close()

	// Create an auth request with a random correlation ID
	correlationID := rand.IntN(1000)
	payload := schemas.AuthEvent{
		EventType: "AuthRequest",
		ID: correlationID,
		AuthRequest: schemas.AuthRequest{
			Username: *user,
			Password: *password,
		},
	}

	// Send a request
	fmt.Printf("Making auth request with correlation ID %d: %+v\n", correlationID, payload)
	if err := makeAuthRequest(ctx, server, payload, parameters.EnvironmentName(*environment)); err != nil {
		fmt.Printf("Error making auth request: %v\n", err)
		os.Exit(1)
	}

	// Await the response with the same correlation ID, ignoring others
	fmt.Printf("Awaiting auth response for correlation ID %d...\n", correlationID)
	if err := awaitAuthResponse(ctx, server, correlationID); err != nil {
		fmt.Printf("Error awaiting auth response: %v\n", err)
		os.Exit(1)
	}
}

func makeAuthRequest(ctx context.Context, producer operations.AuthRequestOperationServerKafka, payload schemas.AuthEvent, env parameters.EnvironmentName) error {
	// Fill out the channel parameters, see AsyncAPI document
	channelParams := channels.AuthChannelParameters{EnvironmentName: env}

	// Open a publisher
	publisher, err := producer.OpenAuthRequestOperationKafka(ctx, channelParams)
	if err != nil {
		return fmt.Errorf("open publisher: %w", err)
	}
	defer publisher.Close()

	// Publish an authorization request
	msg := messages.AuthRequestMsgOut{Payload: payload}
	if err := publisher.PublishAuthRequestMsg(ctx, &msg); err != nil {
		return fmt.Errorf("publish auth request: %w", err)
	}

	return nil
}

func awaitAuthResponse(ctx context.Context, consumer operations.AuthResponseOperationServerKafka, corrID int) error {
	// Fill out the channel parameters, see AsyncAPI document
	channelParams := channels.AuthChannelParameters{EnvironmentName: "development"}

	// Open a subscriber
	subscriber, err := consumer.OpenAuthResponseOperationKafka(ctx, channelParams)
	if err != nil {
		return fmt.Errorf("open subscriber: %w", err)
	}
	defer subscriber.Close()

	awaitCtx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	// Wait for an authorization response with given correlation ID, skip others
	err = subscriber.SubscribeAuthResponseMsg(awaitCtx, func(message messages.AuthResponseMsgReceiver) {
		if message.Payload().EventType != "AuthResponse" {
			fmt.Printf("Ignoring message with event type %s, expected AuthResponse\n", message.Payload().EventType)
			return
		}

		// Check correlation ID
		id, err := message.CorrelationID()
		if err != nil {
			cancel(fmt.Errorf("get correlation ID: %w", err))
			return
		}
		if id != corrID {
			fmt.Printf("Ignoring message with correlation ID %d, expected %d\n", id, corrID)
			return
		}

		fmt.Printf("Received auth response with ID=%v: %+v\n", message.Payload().ID, message.Payload().AuthResponse)
		cancel(nil) // Exit the loop on successful response
	})
	if err != nil && !errors.Is(context.Cause(awaitCtx), context.Canceled) {
		return fmt.Errorf("receive auth response: %w", err)
	}

	return nil
}