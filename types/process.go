package types

import (
	"time"
)

// Persistent data
type ProcessPoint struct {
	ID                  uint      `json:"id"`
	Name                string    `json:"name" gorm:"primaryKey"`
	Description         string    `json:"description"`
	LastValue           float64   `json:"lastvalue"`
	Updated             time.Time `json:"updated"`
	Integrator          float64   `json:"integrator"`
	IntegratingDeadband float64   `json:"integratingdeadband"`
}
