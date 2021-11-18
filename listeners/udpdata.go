package listeners

import (
	"encoding/json"
	"log"
	"math"
	"net"

	"github.com/cyops-se/dd-inserter/db"
	"github.com/cyops-se/dd-inserter/engine"
	"github.com/cyops-se/dd-inserter/types"
)

var prevMsg *types.DataMessage

type UDPDataListener struct {
	Port int `json:"port"`
}

func (listener *UDPDataListener) InitListener() {
	listeners = append(listeners, listener)
	go listener.run()
}

func (listener *UDPDataListener) run() {
	port := 4357
	p := make([]byte, 2048*1024*8)
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("0.0.0.0"),
	}

	ser, err := net.ListenUDP("udp", &addr)
	if err != nil {
		db.Error("UDPDataListener error", "Failed to listen at address: %s, error: %s", addr.String(), err.Error())
		return
	}

	log.Println("UDP listening for DATA messages...")
	for {
		n, _, err := ser.ReadFromUDP(p)
		if err != nil {
			db.Error("UDPDataListener error", "Failed to read UDP data (ReadFromUDP), error: %s", err.Error())
			continue
		}

		var msg types.DataMessage
		if err := json.Unmarshal(p[:n], &msg); err != nil {
			db.Error("UDPDataListener error", "Failed to unmarshal data message, error: %s", err.Error())
			continue
		}

		if prevMsg == nil {
			prevMsg = &msg
		}

		seqdiff := math.Abs(float64(msg.Sequence) - float64(prevMsg.Sequence))
		if seqdiff > 1.0 {
			db.Error("UDPDataListener sequence", "Sequence not valid, now: %d, previous: %d\n", msg.Sequence, prevMsg.Sequence)
			engine.SendAlerts()
		}

		prevMsg = &msg

		// log.Printf("Message received: %s", msg.Points[0].Time)
		engine.NewMsg <- msg
	}
}
