package db

import (
	"fmt"
	"log"
	"time"

	"github.com/cyops-se/dd-inserter/types"
)

func Log(category string, title string, msg string) string {
	entry := &types.Log{Time: time.Now().UTC(), Category: category, Title: title, Description: msg}
	DB.Create(&entry)
	text := fmt.Sprintf("%s: %s, %s", category, title, msg)
	// log.Printf("%s: %s, %s", category, title, msg)
	log.Printf(text)
	return text
}
