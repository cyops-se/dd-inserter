package engine

import (
	"github.com/cyops-se/dd-inserter/db"
	"github.com/cyops-se/dd-inserter/types"
)

func GetAll() ([]types.KeyValuePair, error) {
	var items []types.KeyValuePair
	db.DB.Find(&items)
	return items, db.DB.Error
}

func Get(key string) (string, error) {
	var item types.KeyValuePair
	db.DB.Find(item, "key = ?", key)
	return item.Value, db.DB.Error
}

func Set(key string, value string) error {
	item := &types.KeyValuePair{Key: key, Value: value}
	db.DB.Save(item)
	return nil
}
