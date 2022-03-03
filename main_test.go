package main_test

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/streadway/amqp"
)

func Test1(t *testing.T) {

}

func processSingle(ch *amqp.Channel, q amqp.Queue, beginPort, endPort int) {
	// Attempt to push a message every 2 seconds
	wg := &sync.WaitGroup{}
	go consumeQueue(ch, wg)

	for i := beginPort; i < endPort; i++ {
		wg.Add(1)
		coId := fmt.Sprintf("%d", time.Now().UnixNano())
		addCoId(coId)

		delay := rand.Intn(1000)
		time.Sleep(time.Millisecond * time.Duration(delay))
		body, _ := json.Marshal(makeRequest(i, coId))
		log.Printf(" [x] Sent %s", body)
		log.Printf(" [x] Sent Bytes %d", len(body))
		err := publishMessage(ch, q, coId, body)
		failOnError(err, "Failed to publish a message")
	}

	wg.Wait()
}
