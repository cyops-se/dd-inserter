package types

import "gorm.io/gorm"

type Listener struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Settings    string `json:"settings"`
}

type Emitter struct {
	gorm.Model
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Settings    string `json:"settings"`
}
