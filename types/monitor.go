package types

import "gorm.io/gorm"

type Recipient struct {
	gorm.Model
	Email    string `json:"email"`
	Category string `json:"category"`
}
