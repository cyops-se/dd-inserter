package emitters

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cyops-se/dd-inserter/db"
	"github.com/cyops-se/dd-inserter/types"
	"github.com/streadway/amqp"
)

type RabbitMQEmitter struct {
	Host      string `json:"host"`
	Port      int    `json:"port"`
	User      string `json:"username"`
	Password  string `json:"password"`
	Authident bool   `json:"authident"`
	Database  string `json:"database"`
	Batchsize int    `json:"batchsize"`
}

type RabbitMQDataPoint struct {
	Signal    string    `json:"signal"`
	Value     float64   `json:"value"`
	Status    int       `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

var connection *amqp.Connection
var channel *amqp.Channel
var rqueue amqp.Queue
var err error

func (emitter *RabbitMQEmitter) InitEmitter() {
	emitters = append(emitters, emitter)
	connection, err = amqp.Dial("amqp://admin:hemligt@192.168.0.174:5672/")
	if err != nil {
		db.Log("error", "RabbitMQ init", fmt.Sprintf("Failed to connect RabbitMQ server: %v", err.Error()))
		return
	}

	channel, err = connection.Channel()

	rqueue, err = channel.QueueDeclare(
		"hello", // name
		false,   // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)

	if err != nil {
		db.Log("error", "RabbitMQ init", fmt.Sprintf("Failed to declare queue: %v", err.Error()))
		return
	}
}

func (emitter *RabbitMQEmitter) ProcessMessage(dp *types.DataPoint) {
	if dp == nil {
		return
	}

	// Only accept data points of floating point type
	if _, ok := dp.Value.(float64); !ok {
		return
	}

	// Use safe marshalling to avoid human mistakes when formatting JSON
	rmdp := &RabbitMQDataPoint{Signal: dp.Name, Value: dp.Value.(float64), Status: dp.Quality}
	body, _ := json.Marshal(rmdp)

	err = channel.Publish(
		"",          // exchange
		rqueue.Name, // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			ContentType: "text/json",
			Body:        []byte(body),
		})

	if err != nil {
		db.Log("error", "RabbitMQ init", fmt.Sprintf("Failed to publish message: %v", err.Error()))
		return
	}
}

func (emitter *RabbitMQEmitter) ProcessMeta(dp *types.DataPointMeta) {
}

func (emitter *RabbitMQEmitter) GetStats() *types.EmitterStatistics {
	return nil
}
