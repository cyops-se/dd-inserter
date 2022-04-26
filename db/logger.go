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

	var result int64
	DB.Model(&types.Log{}).Count(&result)
	for result > 1000 {
		var first types.Log
		DB.First(&first)
		DB.Unscoped().Delete(&first)
		result--
	}

	return text
}

func Trace(title string, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	text := Log("trace", title, msg)
	return fmt.Errorf(text)
}

func Error(title string, format string, args ...interface{}) error {
	msg := fmt.Sprintf(format, args...)
	text := Log("error", title, msg)
	return fmt.Errorf(text)
}
