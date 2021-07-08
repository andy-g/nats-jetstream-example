package main

import (
	"encoding/json"
	"log"
	"runtime"

	"github.com/nats-io/nats.go"
)

type Order struct {
	OrderID    int
	CustomerID string
	Status     string
}

const (
	subSubjectName = "ORDERS.created"
	pubSubjectName = "ORDERS.approved"
)

func main() {
	// Connect to NATS
	nc, _ := nats.Connect(nats.DefaultURL)
	js, err := nc.JetStream()
	if err != nil {
		log.Fatal(err)
	}


	// // Create Pull based consumer with maximum 128 inflight.
	// // PullMaxWaiting defines the max inflight pull requests.
	// sub, _ := js.PullSubscribe(subSubjectName, "order-review", nats.PullMaxWaiting(128))
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()

	// for {
	// 	select {
	// 	case <-ctx.Done():
	// 		return
	// 	default:
	// 	}
	// 	msgs, _ := sub.Fetch(10, nats.Context(ctx))
	// 	for _, msg := range msgs {
	// 		msg.Ack()
	// 		var order Order
	// 		err := json.Unmarshal(msg.Data, &order)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		log.Println("order-review service")
	// 		log.Printf("OrderID:%d, CustomerID: %s, Status:%s\n", order.OrderID, order.CustomerID, order.Status)
	// 		reviewOrder(js, order)
	// 	}
	// }

	// Create durable consumer monitor
	js.Subscribe(subSubjectName, func(msg *nats.Msg) {
		msg.Ack()
		var order Order
		err := json.Unmarshal(msg.Data, &order)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("OrderID:%d, CustomerID: %s, Status:%s\n", order.OrderID, order.CustomerID, order.Status)
		reviewOrder(js, order)
	}, nats.Durable("order-review"), nats.ManualAck())

	runtime.Goexit()
}

// reviewOrder reviews the order and publishes ORDERS.approved event
func reviewOrder(js nats.JetStreamContext, order Order) {
	// Changing the Order status
	order.Status = "approved"
	orderJSON, _ := json.Marshal(order)
	_, err := js.Publish(pubSubjectName, orderJSON)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Order with OrderID:%d has been %s\n", order.OrderID, order.Status)
}