package listeners

import (
	"encoding/json"
	"log"
	"net"

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
		log.Printf("Failed to listen %v\n", err)
		return
	}

	log.Println("UDP listening for DATA messages...")
	for {
		n, _, err := ser.ReadFromUDP(p)
		if err != nil {
			log.Printf("Some error  %v", err)
			continue
		}

		var msg types.DataMessage
		if err := json.Unmarshal(p[:n], &msg); err != nil {
			log.Println("Failed to unmarshal data, err:", err)
			continue
		}

		if prevMsg == nil {
			prevMsg = &msg
		}

		if msg.Counter-prevMsg.Counter > 1 {
			log.Printf("ERROR sequence not valid, now: %d, previous: %d\n", msg.Counter, prevMsg.Counter)
		}

		prevMsg = &msg

		// log.Printf("Message received: %s", msg.Points[0].Time)
		engine.NewMsg <- msg
	}
}
