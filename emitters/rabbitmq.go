package emitters

import (
	"encoding/json"
	"fmt"

	"github.com/cyops-se/dd-inserter/db"
	"github.com/cyops-se/dd-inserter/types"
	"github.com/streadway/amqp"
)

type RabbitMQEmitter struct {
	EmitterBase
	Host      string `json:"host"`
	Port      int    `json:"port"`
	User      string `json:"username"`
	Password  string `json:"password"`
	Authident bool   `json:"authident"`
	Database  string `json:"database"`
	Batchsize int    `json:"batchsize"`
}

var connection *amqp.Connection
var channel *amqp.Channel
var rqueue amqp.Queue
var err error

func (emitter *RabbitMQEmitter) InitEmitter() {
	fmt.Println("RABBITMQ emitter processing message")
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

	emitters = append(emitters, emitter)
}

func (emitter *RabbitMQEmitter) ProcessMessage(dp *types.DataPoint) {
	if dp == nil {
		return
	}

	// defer connection.Close()
	// defer channel.Close()

	// fmt.Println("RABBITMQ emitter processing message")

	body, _ := json.Marshal(dp)
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
