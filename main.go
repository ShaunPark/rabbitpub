package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/streadway/amqp"
	klog "k8s.io/klog/v2"
)

func main() {
	name := "testqueue"
	conn, err := amqp.Dial("amqp://username:password@localhost:5672/")
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		name,  // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	i := 0
	s2 := rand.NewSource(int64(time.Now().Second()))
	r2 := rand.New(s2)
	cnt := r2.Intn(20)
	log.Print(cnt)
	// Attempt to push a message every 2 seconds
	wg := sync.WaitGroup{}

	go func() {
		{
			q, err := ch.QueueDeclare(
				reply_queue, // name
				true,        // durable
				false,       // delete when unused
				false,       // exclusive
				false,       // noWait
				nil,         // arguments
			)
			if err != nil {
				klog.Error(err, "Failed to declare a queue")
			}
			msgs, err := ch.Consume(
				q.Name, // queue
				"",     // consumer
				true,   // auto-ack
				false,  // exclusive
				false,  // no-local
				false,  // no-wait
				nil,    // args
			)
			if err != nil {
				klog.Error(err, "Failed to register a consumer")
			}

			for msg := range msgs {
				klog.Info(msg)
				wg.Done()
			}
		}
	}()

	for i < cnt {
		wg.Add(1)
		delay := rand.Intn(1000)
		time.Sleep(time.Millisecond * time.Duration(delay))
		body, _ := json.Marshal(makeRequest(10000 + i))
		i++

		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        body,
			})
		failOnError(err, "Failed to publish a message")
		log.Printf(" [x] Sent %s", body)
	}

	wg.Wait()
}

func makeRequest(port int) BeeRequest {
	return BeeRequest{
		MetaData: &MetaData{
			Type:          "ADD",
			SubType:       "",
			From:          "Tester",
			To:            "proxyUpdater",
			Queue:         reply_queue,
			CorrelationId: fmt.Sprintf("%d", time.Millisecond),
		},
		PayLoad: RequestPayLoad{
			RequestName: "create",
			Data: RequestData{
				WorkerId:   fmt.Sprintf("worker-%d", port),
				ClusterIps: []string{"10.0.0.12"},
				NodeIp:     "10.0.0.12",
				NodePort:   port + 10000,
				ProxyPort:  port,
			},
		},
	}
}
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

const (
	reply_queue = "rep_queue"
)
