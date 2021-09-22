package listeners

import (
	"gorm.io/gorm"
)

type Listener interface {
	InitListener()
}

type ListenerBase struct {
	gorm.Model
	Name             string `json:"name"`
	Type             string `json:"type"`
	MessagePerSecond uint64 `json:"mps"`
}

var listeners []*ListenerBase

func RunDispatch() {
}

func (listener *ListenerBase) RegisterListener() {
	listeners = append(listeners, listener)
}
