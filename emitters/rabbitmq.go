package emitters

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/cyops-se/dd-inserter/db"
	"github.com/cyops-se/dd-inserter/types"
	"github.com/sirius1024/go-amqp-reconnect/rabbitmq"
	"github.com/streadway/amqp"
)

type RabbitMQEmitter struct {
	// The attributes below are serialized into the 'Settings' attribute of the Emitter attribute above
	Urls        []string `json:"urls"`
	ChannelName string   `json:"channel"`
	connection  *rabbitmq.Connection
	channel     *rabbitmq.Channel
	queue       amqp.Queue
	err         error
	initialized bool
}

type RabbitMQDataPoint struct {
	Signal    string    `json:"signal"`
	Value     float64   `json:"value"`
	Status    int       `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

func (emitter *RabbitMQEmitter) InitEmitter() error {
	if len(emitter.Urls) == 0 {
		emitter.err = fmt.Errorf("Failed to connect RabbitMQ cluster, urls parameter empty")
		db.Log("error", "RabbitMQ init", emitter.err.Error())
		return emitter.err
	}

	emitter.connection, emitter.err = rabbitmq.DialCluster(emitter.Urls)
	if emitter.err != nil {
		db.Log("error", "RabbitMQ init", fmt.Sprintf("Failed to connect RabbitMQ cluster [%s]: %v", emitter.Urls, emitter.err.Error()))
		return emitter.err
	}

	emitter.channel, emitter.err = emitter.connection.Channel()

	emitter.queue, emitter.err = emitter.channel.QueueDeclare(
		emitter.ChannelName, // name
		false,               // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)

	if emitter.err != nil {
		db.Log("error", "RabbitMQ init", fmt.Sprintf("Failed to declare queue: %v", emitter.err.Error()))
		return emitter.err
	}

	emitter.initialized = true
	return emitter.err
}

// func (emitter *RabbitMQEmitter) LoadSettingsJSON(settings string) error {
// 	// settings is a JSON object with all settings (serialized from RabbitMQEmitter)
// 	return json.Unmarshal([]byte(settings), &emitter)
// }

// func (emitter *RabbitMQEmitter) GetSettingsJSON() (string, error) {
// 	settings, err := json.Marshal(emitter)
// 	if err != nil {
// 		db.Log("error", "Failed to save RabbitMQ settings", err.Error())
// 		return "", err
// 	}

// 	return string(settings), nil
// }

func (emitter *RabbitMQEmitter) ProcessMessage(dp *types.DataPoint) {
	if dp == nil || !emitter.initialized {
		return
	}

	// Only accept data points of floating point type
	if _, ok := dp.Value.(float64); !ok {
		return
	}

	// Use safe marshalling to avoid human mistakes when formatting JSON
	rmdp := &RabbitMQDataPoint{Signal: dp.Name, Value: dp.Value.(float64), Status: dp.Quality}
	body, _ := json.Marshal(rmdp)

	emitter.err = emitter.channel.Publish(
		"",                 // exchange
		emitter.queue.Name, // routing key
		false,              // mandatory
		false,              // immediate
		amqp.Publishing{
			ContentType: "text/json",
			Body:        []byte(body),
		})

	if emitter.err != nil {
		db.Log("error", "RabbitMQ init", fmt.Sprintf("Failed to publish message: %v", emitter.err.Error()))
		return
	}
}

func (emitter *RabbitMQEmitter) ProcessMeta(dp *types.DataPointMeta) {
}

func (emitter *RabbitMQEmitter) GetStats() *types.EmitterStatistics {
	return nil
}
