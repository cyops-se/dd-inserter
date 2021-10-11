package types

import "gorm.io/gorm"

type IEmitter interface {
	InitEmitter() error
	ProcessMessage(msg *DataPoint)
	ProcessMeta(msg *DataPointMeta)
	GetStats() *EmitterStatistics
}

type Emitter struct {
	gorm.Model
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Type        string   `json:"type"`
	Settings    string   `json:"settings"`
	Instance    IEmitter `json:"instance" gorm:"-"`
}
