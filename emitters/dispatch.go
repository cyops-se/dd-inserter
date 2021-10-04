package emitters

import (
	"github.com/cyops-se/dd-inserter/engine"
	"github.com/cyops-se/dd-inserter/types"
)

type IEmitter interface {
	InitEmitter()
	ProcessMessage(msg *types.DataPoint)
	ProcessMeta(msg *types.DataPointMeta)
	GetStats() *types.EmitterStatistics
}

var emitters []IEmitter

func RunDispatch() {
	go messageDispatch()
	go metaDispatch()
}

func Init() {
	ts := &TimescaleEmitter{}
	ts.InitEmitter()

	rmq := &RabbitMQEmitter{}
	rmq.InitEmitter()
}

func messageDispatch() {
	for {
		msg := <-engine.NewEmitMsg
		for _, emitter := range emitters {
			emitter.ProcessMessage(&msg)
		}

		// Always emit to websocket subscribers
		engine.NotifySubscribers("data.message", msg)
	}
}

func metaDispatch() {
	for {
		msg := <-engine.NewEmitMetaMsg
		for _, emitter := range emitters {
			emitter.ProcessMeta(&msg)
		}

		// Always emit to websocket subscribers
		engine.NotifySubscribers("data.meta", msg)
	}
}
