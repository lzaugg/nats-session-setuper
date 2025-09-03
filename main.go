package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"nats-setuper/internal"
	"os"

	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

func main() {
	err := run()
	if err != nil {
		slog.Error("error running", "error", err.Error())
		os.Exit(1)
	}
}

func run() error {

	envFile := flag.String("env-file", ".env", "env file to load credentials from (default: .env)")
	flag.Parse()

	var server string

	err := godotenv.Load(*envFile)
	if err != nil {
		return fmt.Errorf("error loading .env file. Maybe inception? Just create a .env file on your own with the correct NATS_SERVER in it (nats://<user>:<token>@<server>:<port>): %w", err)
	}

	server = os.Getenv("NATS_SERVER")

	if server == "" {
		return fmt.Errorf("no server info found in .env file")
	}

	nc, err := nats.Connect(server)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS: %w", err)
	}
	defer nc.Close()

	ctxt := context.Background()
	ac, err := internal.NewAtomicCounter(ctxt, nc, "atomic_counter", "last_user_id")
	if err != nil {
		return fmt.Errorf("failed to create atomic counter: %w", err)
	}

	srv, err := micro.AddService(nc, micro.Config{
		Name:    "GopherService",
		Version: "1.0.0",
	})

	if err != nil {
		return fmt.Errorf("failed to create micro service: %w", err)
	}
	err = srv.AddEndpoint(
		"next-gopher",
		micro.ContextHandler(ctxt, createUserNameHandler(ac)),
		micro.WithEndpointSubject("service.next-gopher"),
	)
	if err != nil {
		return fmt.Errorf("failed to create micro endpoint: %w", err)
	}

	err = srv.AddEndpoint(
		"ping",
		micro.ContextHandler(ctxt, createPingHandler()),
		micro.WithEndpointSubject("service.ping"),
	)
	if err != nil {
		return fmt.Errorf("failed to create micro endpoint: %w", err)
	}

	slog.Info("micro service created")

	select {}

}

func createUserNameHandler(ac *internal.AtomicCounter) func(ctxt context.Context, req micro.Request) {
	return func(ctxt context.Context, req micro.Request) {
		userId, err := ac.GetNextValue(ctxt)
		if err != nil {
			slog.Error("failed to get next user id", "error", err)
			req.Error("500", "Cannot get next user id", []byte(err.Error()))
			return
		}
		userName := fmt.Sprintf("gopher-%02d", userId)
		req.Respond([]byte(userName))
	}
}

func createPingHandler() func(ctxt context.Context, req micro.Request) {
	return func(ctxt context.Context, req micro.Request) {
		req.Respond([]byte("pong"))
	}
}
