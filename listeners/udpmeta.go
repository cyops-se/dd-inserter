package listeners

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/cyops-se/dd-inserter/engine"
	"github.com/cyops-se/dd-inserter/types"
)

type UDPMetaListener struct {
	Port int `json:"port"`
}

func (listener *UDPMetaListener) InitListener() {
	listeners = append(listeners, listener)
	go listener.run()
}

func (listener *UDPMetaListener) run() {
	port := 4356
	p := make([]byte, 12048)
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("0.0.0.0"),
	}

	ser, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("Failed to listen %v\n", err)
		return
	}

	log.Println("UDP listening for META messages...")
	for {
		n, _, err := ser.ReadFromUDP(p)
		if err != nil {
			fmt.Printf("Some error  %v", err)
			continue
		}

		var msg []*types.DataPointMeta
		if err := json.Unmarshal(p[:n], &msg); err != nil {
			fmt.Println("Failed to unmarshal data, err:", err)
			return
		}

		engine.NewMeta <- msg
	}
}
