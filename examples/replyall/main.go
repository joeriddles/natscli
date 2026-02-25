package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/micro"
)

func main() {
	// start embedded nats server
	srv, err := server.NewServer(&server.Options{
		Port: -1,
	})
	if err != nil {
		log.Fatalf("could not start server: %v", err)
	}

	srv.Start()
	if !srv.ReadyForConnections(3 * time.Second) {
		log.Fatalf("nats server did not start")
	}

	defer srv.Shutdown()

	// create connection to nats server
	nc, err := nats.Connect(srv.ClientURL())
	if err != nil {
		log.Fatalf("could not connect to server: %v", err)
	}
	defer nc.Close()

	// start micro service
	svc, err := micro.AddService(nc, micro.Config{
		Name:    "svc",
		Version: "1.0.0",
	})
	if err != nil {
		log.Fatalf("failed to add service: %v", err)
	}
	defer svc.Stop()

	err = svc.AddEndpoint("reply", micro.HandlerFunc(func(req micro.Request) {
		// random number between 1 and 100
		rng := rand.Intn(100) + 1

		// send three replies
		headers := nats.Header{"NATS-Replies-Total": {strconv.Itoa(rng)}}
		for i := range rng {
			req.Respond([]byte(fmt.Sprintf("Response %d", i)), micro.WithHeaders(micro.Headers(headers)))
		}
	}))
	if err != nil {
		log.Fatalf("failed to add service endpoint reply: %v", err)
	}

	fmt.Printf("started service on %s\n", srv.ClientURL())

	// run until interrupted
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
}
