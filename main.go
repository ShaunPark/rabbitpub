package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ShaunPark/rabbitPub/types"
	"github.com/ShaunPark/rabbitPub/utils"
	"github.com/streadway/amqp"
	"gopkg.in/alecthomas/kingpin.v2"
	klog "k8s.io/klog/v2"
)

const (
	reply_queue = "rep_queue"
)

var (
	configFile = kingpin.Flag("configFile", "configFile").Short('f').Required().String()
	mode       = kingpin.Flag("mode", "log level").Short('m').Required().String()
	ports      = kingpin.Flag("ports", "end of port range").Short('p').String()

	config *types.Config

	coIds = map[string]string{}

	mutex = sync.Mutex{}
)

// read config file
func getConfig() *types.Config {
	configManager := utils.NewConfigManager(*configFile)
	config := configManager.GetConfig()
	if config.TestMode {
		config.RabbitMQ.Server.Password = "password"
		config.HaProxy.Master.Password = "adminpwd"
		config.HaProxy.Second.Password = "adminpwd"
	} else {
		config.RabbitMQ.Server.Password = os.Getenv(utils.ENV_RABBIT_MQ_PWD)
		config.HaProxy.Master.Password = os.Getenv(utils.ENV_HAPROXY_PWD)
		config.HaProxy.Second.Password = os.Getenv(utils.ENV_HAPROXY_PWD)
	}
	return config
}

func publishMessage(ch *amqp.Channel, q amqp.Queue, coId string, body []byte) error {
	return ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType:   "text/plain",
			ReplyTo:       reply_queue,
			CorrelationId: coId,
			Body:          body,
		})

}

func addCoId(id string) {
	mutex.Lock()
	defer mutex.Unlock()
	coIds[id] = id
}

func processBatch(ch *amqp.Channel, q amqp.Queue, beginPort, endPort int) {
	wg := &sync.WaitGroup{}
	coId := fmt.Sprintf("%d", time.Now().UnixNano())

	go consumeQueue(ch, wg)
	wg.Add(1)
	addCoId(coId)

	body, _ := json.Marshal(makeBatchRequest(coId, beginPort, endPort, "BATCH"))
	err := publishMessage(ch, q, coId, body)
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s", body)
	wg.Wait()
}

func consumeQueue(ch *amqp.Channel, wg *sync.WaitGroup) {
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
		rep := &types.BeeResponse{}
		json.Unmarshal(msg.Body, rep)
		klog.Info(rep)
		coId := rep.MetaData.CorrelationId

		go func() {
			mutex.Lock()
			defer mutex.Unlock()
			if _, exist := coIds[coId]; exist {
				delete(coIds, coId)
				wg.Done()
			}
		}()
	}
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

		err := publishMessage(ch, q, coId, body)
		failOnError(err, "Failed to publish a message")
		log.Printf(" [x] Sent %s", body)
	}

	wg.Wait()
}

func processTest(ch *amqp.Channel, q amqp.Queue, beginPort, endPort int) {
	wg := &sync.WaitGroup{}
	coId := fmt.Sprintf("%d", time.Now().UnixNano())

	go consumeQueue(ch, wg)
	wg.Add(1)
	addCoId(coId)

	body, _ := json.Marshal(makeBatchRequest(coId, beginPort, endPort, "BATCHLOAD"))
	err := publishMessage(ch, q, coId, body)
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s", body)
	wg.Wait()
}

func processBatchDelete(ch *amqp.Channel, q amqp.Queue, beginPort, endPort int) {
	wg := &sync.WaitGroup{}
	coId := fmt.Sprintf("%d", time.Now().UnixNano())

	go consumeQueue(ch, wg)
	wg.Add(1)
	addCoId(coId)

	body, _ := json.Marshal(makeBatchRequest(coId, beginPort, endPort, "BATCHDELETE"))
	err := publishMessage(ch, q, coId, body)
	failOnError(err, "Failed to publish a message")
	log.Printf(" [x] Sent %s", body)
	wg.Wait()
}
func processDelete(ch *amqp.Channel, q amqp.Queue, beginPort, endPort int) {
	// Attempt to push a message every 2 seconds
	wg := &sync.WaitGroup{}
	go consumeQueue(ch, wg)

	for i := beginPort; i < endPort; i++ {
		wg.Add(1)
		coId := fmt.Sprintf("%d", time.Now().UnixNano())
		addCoId(coId)
		body, _ := json.Marshal(makeDeleteRequest(i, coId))

		err := publishMessage(ch, q, coId, body)
		failOnError(err, "Failed to publish a message")
		log.Printf(" [x] Sent %s", body)
	}

	wg.Wait()
}

func main() {
	kingpin.Parse()
	config = getConfig()
	mo := *mode

	name := config.RabbitMQ.Queue
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%s/", config.RabbitMQ.Server.Id, config.RabbitMQ.Server.Password, config.RabbitMQ.Server.Host, config.RabbitMQ.Server.Port))
	failOnError(err, "Failed  to connect to RabbitMQ")
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
	ps := strings.Split(*ports, ":")
	ep, err := strconv.Atoi(ps[1])
	failOnError(err, "endPort parse error")
	bp, err := strconv.Atoi(ps[0])
	failOnError(err, "begin parse error")

	switch mo {
	case "s":
		processSingle(ch, q, bp, ep)
	case "b":
		processBatch(ch, q, bp, ep)
	case "d":
		processDelete(ch, q, bp, ep)
	case "t":
		processTest(ch, q, bp, ep)
	case "r":
		processBatchDelete(ch, q, bp, ep)
	}
}

func makeRequest(port int, coId string) types.BeeRequest {
	return types.BeeRequest{
		MetaData: &types.MetaData{
			Type:          "NETWORK_CFG",
			SubType:       "ADD",
			From:          "Tester",
			To:            "HAProxyUpdater",
			Queue:         "testqueue",
			CorrelationId: coId,
		},
		PayLoad: &types.RequestPayLoad{
			RequestName: "createWorker",
			Data: &types.RequestData{
				WorkerId:   fmt.Sprintf("worker-%d", port),
				ClusterIps: &[]string{"10.0.0.12"},
				NodeIp:     "10.0.0.12",
				NodePort:   port + 20000,
				ProxyPort:  port + 10000,
			},
		},
	}
}

func makeDeleteRequest(port int, coId string) types.BeeRequest {
	return types.BeeRequest{
		MetaData: &types.MetaData{
			Type:          "NETWORK_CFG",
			SubType:       "DELETE",
			From:          "Tester",
			To:            "HAProxyUpdater",
			Queue:         "testqueue",
			CorrelationId: coId,
		},
		PayLoad: &types.RequestPayLoad{
			RequestName: "deleteWorker",
			Data: &types.RequestData{
				WorkerId:  fmt.Sprintf("worker-%d", port),
				ProxyPort: port + 10000,
			},
		},
	}
}

// func makeTestRequest(coId string) types.BeeRequest {
// 	return types.BeeRequest{
// 		MetaData: &types.MetaData{
// 			Type:          "NETWORK_CFG",
// 			SubType:       "BATCHLOAD",
// 			From:          "Tester",
// 			To:            "HAProxyUpdater",
// 			Queue:         "testqueue",
// 			CorrelationId: coId,
// 		},
// 		PayLoad: types.RequestPayLoad{
// 			RequestName: "deleteWorker",
// 			Data: types.RequestData{
// 				WorkerId:  fmt.Sprintf("worker-%d", 1),
// 				ProxyPort: 1 + 10000,
// 			},
// 		},
// 	}
// }

func makeBatchRequest(coId string, beginPort, endPort int, typeStr string) types.BeeRequest {
	data := []types.RequestData{}

	for i := beginPort; i < endPort; i++ {
		if typeStr == "BATCHDELETE" {
			data = append(data, types.RequestData{
				WorkerId:  fmt.Sprintf("worker-%d", i),
				ProxyPort: i + 10000,
			})
		} else {
			data = append(data, types.RequestData{
				WorkerId:   fmt.Sprintf("worker-%d", i),
				ClusterIps: &[]string{"10.0.0.12"},
				NodeIp:     "10.0.0.12",
				NodePort:   i + 20000,
				ProxyPort:  i + 10000,
			})
		}
	}
	return types.BeeRequest{
		MetaData: &types.MetaData{
			Type:          "NETWORK_CFG",
			SubType:       typeStr,
			From:          "Tester",
			To:            "HAProxyUpdater",
			Queue:         "testqueue",
			CorrelationId: coId,
		},
		PayLoad: &types.RequestPayLoad{
			RequestName: "createWorker",
			BatchData:   &data,
		},
	}
}
func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}
