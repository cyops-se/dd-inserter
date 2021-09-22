package emitters

import (
	"github.com/cyops-se/dd-inserter/engine"
	"github.com/cyops-se/dd-inserter/types"
	"gorm.io/gorm"
)

type Emitter interface {
	InitEmitter()
	ProcessMessage(msg *types.DataPoint)
}

type EmitterBase struct {
	gorm.Model
	Name string `json:"name"`
	Type string `json:"type"`
}

var emitters []Emitter

func RunDispatch() {
	for {
		msg := <-engine.NewEmitMsg
		for _, emitter := range emitters {
			emitter.ProcessMessage(&msg)
		}

		// Always emit to websocket subscribers
		engine.NotifySubscribers("data.message", msg)
	}
}
