package types

import "gorm.io/gorm"

type IListener interface {
	InitListener()
}

type Listener struct {
	gorm.Model
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Settings    string    `json:"settings"`
	Instance    IListener `gorm:"-"`
}
