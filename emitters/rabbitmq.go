package emitters

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/cyops-se/dd-inserter/db"
	"github.com/cyops-se/dd-inserter/types"
	"github.com/sirius1024/go-amqp-reconnect/rabbitmq"
	"github.com/streadway/amqp"
)

type RabbitMQEmitter struct {
	// The attributes below are serialized into the 'Settings' attribute of the Emitter attribute above
	Urls        []string              `json:"urls"`
	ChannelName string                `json:"channel"`
	Durable     bool                  `json:"durable"`
	connection  *rabbitmq.Connection  `json:"-"`
	channel     *rabbitmq.Channel     `json:"-"`
	queue       amqp.Queue            `json:"-"`
	err         error                 `json:"-"`
	initialized bool                  `json:"-"`
	messages    chan *types.DataPoint `json:"-"`
}

type RabbitMQDataPoint struct {
	Signal    string    `json:"signal"`
	Value     float64   `json:"value"`
	Status    int       `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type RabbitMQMetaItem struct {
	Signal      string  `json:"signal"`
	Description string  `json:"description"`
	Dimension   string  `json:"dimension"`
	Min         float64 `json:"range_min"`
	Max         float64 `json:"range_max"`
	Deadband    float64 `json:"deadband"`
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
		emitter.Durable,     // durable
		false,               // delete when unused
		false,               // exclusive
		false,               // no-wait
		nil,                 // arguments
	)

	if emitter.err != nil {
		db.Log("error", "RabbitMQ init", fmt.Sprintf("Failed to declare queue: %v", emitter.err.Error()))
		return emitter.err
	}

	emitter.messages = make(chan *types.DataPoint, 2000)
	go emitter.processMessages()
	go emitter.syncMetaRabbit()

	emitter.initialized = true
	db.Log("info", "RABBITMQ emitter", fmt.Sprintf("RabbitMQ server connected: %s", emitter.Urls))
	return nil
}

func (emitter *RabbitMQEmitter) ProcessMessage(dp *types.DataPoint) {
	if dp == nil || !emitter.initialized {
		return
	}

	emitter.messages <- dp
}

func (emitter *RabbitMQEmitter) processMessages() {
	for {
		dp := <-emitter.messages

		switch v := dp.Value.(type) {
		case float64: // Skip
		case int:
			dp.Value = float64(v)
		case uint:
			dp.Value = float64(v)
		case int64:
			dp.Value = float64(v)
		case uint64:
			dp.Value = float64(v)
		case float32:
			dp.Value = float64(v)
		default:
			db.Log("error", "RabbitMQ emitter", fmt.Sprintf("Datapoint '%s' has an unsupported type: '%T'", dp.Name, dp.Value))
			continue
		}

		// Use safe marshalling to avoid human mistakes when formatting JSON
		rmdp := &RabbitMQDataPoint{Signal: dp.Name, Value: dp.Value.(float64), Status: dp.Quality, Timestamp: dp.Time}
		body, _ := json.Marshal(rmdp)

		emitter.err = emitter.channel.Publish(
			"",                 // exchange
			emitter.queue.Name, // routing key
			false,              // mandatory
			false,              // immediate
			amqp.Publishing{
				ContentType: "text/json",
				Body:        body,
			})

		if emitter.err != nil {
			db.Log("error", "RabbitMQ emitter", fmt.Sprintf("Failed to publish message: %v (processMessages)", emitter.err.Error()))
			continue
		}

		// db.Log("info", "RabbitMQ emitter", fmt.Sprintf("Message published: %s (processMessages)", string(body)))
	}
}

// Implement IEmitter interface
func (emitter *RabbitMQEmitter) ProcessMeta(dp *types.DataPointMeta) {
}

// Implement IEmitter interface
func (emitter *RabbitMQEmitter) GetStats() *types.EmitterStatistics {
	return nil
}

func (emitter *RabbitMQEmitter) syncMetaRabbit() {
	ticker := time.NewTicker(30 * time.Second)
	var prevmetaitems []types.DataPointMeta
	var dosend bool
	for {
		<-ticker.C
		var metaitems []types.DataPointMeta
		if err := db.DB.Find(&metaitems).Error; err != nil {
			fmt.Println("TIMESCALE failed to get meta items,", err.Error())
			continue
		}

		// Check individual items (discard items no longer in db)
		for _, dp := range metaitems {
			dosend = true
			for _, pdp := range prevmetaitems {
				if reflect.DeepEqual(dp, pdp) {
					dosend = false
					continue
				}
			}

			if dosend {
				emitter.sendMetaRabbit(&dp)
			}
		}

		prevmetaitems = metaitems
	}
}

func (emitter *RabbitMQEmitter) sendMetaRabbit(dp *types.DataPointMeta) {
	msg := &RabbitMQMetaItem{}
	msg.Signal = dp.Name
	msg.Description = dp.Description
	msg.Dimension = dp.EngUnit
	msg.Max = dp.MaxValue
	msg.Min = dp.MinValue
	msg.Deadband = dp.IntegratingDeadband

	body, _ := json.Marshal(msg)

	emitter.err = emitter.channel.Publish(
		"",            // exchange
		"me-metadata", // routing key
		false,         // mandatory
		false,         // immediate
		amqp.Publishing{
			ContentType: "text/json",
			Body:        body,
		})

	if emitter.err != nil {
		db.Log("error", "RabbitMQ emitter", fmt.Sprintf("Failed to publish meta message: %v", emitter.err.Error()))
	}
}
