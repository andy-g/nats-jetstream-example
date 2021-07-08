package main

import (
	"fmt"
	"log"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	servers     = "nats://localhost:4222"
	streamName  = "ORDERS"
	subjectName = "ORDERS.received"
)

func main() {
	log.Print("connecting securely to cluster")
	nc, err := connect()
	noerr(err)
	defer nc.Close()

	log.Print("getting JetStream context")
	js, err := nc.JetStream()
	noerr(err)

	// Check if the ORDERS stream already exists; if not, create it.
	stream, err := js.StreamInfo(streamName)
	if err != nil {
		log.Print(err)
	}
	if stream == nil {
		log.Printf("creating stream %q and subject %q", streamName, subjectName)
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     streamName,
			Subjects: []string{subjectName},
		})
		noerr(err)
	}

	log.Print("publishing an order")
	_, err = js.Publish(subjectName, []byte("one big burger - 1"))
	noerr(err)

	log.Print("attempting to receive order")
	var order []byte
	done := make(chan bool, 1)
	js.Subscribe(subjectName, func(m *nats.Msg) {
		order = m.Data
		m.Ack()
		done <- true
	})

	select {
	case <-time.After(5 * time.Second):
		log.Fatalf("failed to get order")
	case <-done:
		log.Printf("got order: %q", order)
	}

}

// connect to a JetStream cluster using X509 certificates to authenticate securely.
func connect() (*nats.Conn, error) {
	nc, err := nats.Connect(servers)
	if err != nil {
		return nil, fmt.Errorf("Got an error on Connect with Secure Options: %s\n", err)
	}
	return nc, nil
}

func noerr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
