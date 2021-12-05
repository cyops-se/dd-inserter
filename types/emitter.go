package types

import (
	"time"

	"gorm.io/gorm"
)

type IEmitter interface {
	InitEmitter() error
	ProcessMessage(msg *DataPoint)
	ProcessMeta(msg *DataPointMeta)
	GetStats() *EmitterStatistics
}

type Emitter struct {
	gorm.Model
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Settings    string    `json:"settings"`
	Instance    IEmitter  `json:"instance" gorm:"-"`
	Count       uint64    `json:"count"`
	LastRun     time.Time `json:"lastrun"`
	Status      int       `json:"status"` // 0 = stopped, 1 = running
}
