package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/streadway/amqp"
)

func main() {
	name := "testqueue"
	conn, err := amqp.Dial("amqp://username:password@3.37.120.95:5672/")
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
	for i < cnt {
		delay := rand.Intn(1000)
		time.Sleep(time.Millisecond * time.Duration(delay))
		body := fmt.Sprintf("%d-%s", i, "message")
		i++

		err = ch.Publish(
			"",     // exchange
			q.Name, // routing key
			false,  // mandatory
			false,  // immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(body),
			})
		failOnError(err, "Failed to publish a message")
		log.Printf(" [x] Sent %s", body)
	}

}
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
