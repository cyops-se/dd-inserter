package listeners

import (
	"encoding/json"
	"log"
	"math"

	"github.com/cyops-se/dd-inserter/db"
	"github.com/cyops-se/dd-inserter/engine"
	"github.com/cyops-se/dd-inserter/types"
	"github.com/nats-io/nats.go"
)

var prevNatsDataMsg *types.DataMessage

type NATSDataListener struct {
	URL string `json:"url"`
	nc  *nats.Conn
}

func (listener *NATSDataListener) InitListener(ctx *types.Context) (err error) {
	listener.nc, err = nats.Connect(nats.DefaultURL)
	if err == nil {
		listener.nc.Subscribe("data", listener.callbackHandler)
		log.Printf("NATS server connected")
	}
	return err
}

func (listener *NATSDataListener) callbackHandler(natsmsg *nats.Msg) {

	var msg types.DataMessage
	if err := json.Unmarshal(natsmsg.Data, &msg); err != nil {
		db.Error("NATSDataListener error", "Failed to unmarshal data message, error: %s", err.Error())
		log.Println(string(natsmsg.Data))
		return
	}

	if prevNatsDataMsg == nil {
		prevNatsDataMsg = &msg
	}

	seqdiff := math.Abs(float64(msg.Sequence) - float64(prevNatsDataMsg.Sequence))
	if seqdiff > 1.0 {
		// db.Error("NATSDataListener sequence", "Sequence not valid, now: %d, previous: %d\n", msg.Sequence, prevNatsDataMsg.Sequence)
		engine.SendAlerts()
	}

	prevNatsDataMsg = &msg

	// log.Printf("Message received: %s", msg.Points[0].Time)
	engine.NewMsg <- msg
}
