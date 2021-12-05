package types

import (
	"time"
)

const (
	GroupStatusNotReceiving = iota
	GroupStatusReceiving    = iota
)

const (
	UpdateTypePassthru = iota
	UpdateTypeInterval = iota
	UpdateTypeDeadband = iota
	UpdateTypeDevNull  = iota
)

// Version 2 message types
type DataPoint struct {
	ID      int         `json:"id"`
	Time    time.Time   `json:"t"`
	Name    string      `json:"n"`
	Value   interface{} `json:"v"`
	Quality int         `json:"q"`
}

type DataMessage struct {
	Version  int         `json:"version"`
	Group    string      `json:"group"`
	Interval int         `json:"interval"`
	Sequence uint64      `json:"sequence"`
	Count    int         `json:"count"`
	Points   []DataPoint `json:"points"`
}

type DataPointMeta struct {
	ID                  uint      `json:"id"`
	Name                string    `json:"name" gorm:"primaryKey"`
	Description         string    `json:"description"`
	EngUnit             string    `json:"engunit"`
	MinValue            string    `json:"minvalue"`
	MaxValue            string    `json:"maxvalue"`
	Quantity            string    `json:"quantity"`
	UpdateType          int       `json:"updatetype"` // 0 = pass thru, 1 = interval, 2 = integrating deadband
	Interval            int       `json:"interval"`
	Integrator          float64   `json:"integrator"`
	IntegratingDeadband float64   `json:"integratingdeadband"`
	LastEmitted         time.Time `json:"lastemitted"`
}

type VolatileDataPoint struct {
	DataPoint           *DataPoint `json:"datapoint"`
	StoredValue         float64    `json:"storedvalue"` // only used for the integrating deadband (floating data points)
	Integrator          float64    `json:"integrator"`
	LastEmitted         time.Time  `json:"lastemitted"`
	UpdateType          int        `json:"updatetype"` // 0 = pass thru, 1 = interval, 2 = integrating deadband
	Interval            int        `json:"interval"`
	IntegratingDeadband float64    `json:"integratingdeadband"`
}
