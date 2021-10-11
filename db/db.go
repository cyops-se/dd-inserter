package db

import (
	"log"

	"github.com/cyops-se/dd-inserter/types"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDatabase() {
	database, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	// dsn := "user=dev password=hemligt dbname=dev host=localhost port=5432"
	// database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Println("Failed to connect to local database", err)
		return
	}

	log.Println("Local database connected!")

	// The User model is special due to the 'password' field, and has
	// user specific routes
	database.AutoMigrate(&types.User{})

	// Generic CRUD data types
	configureTypes(database, types.Log{}, types.KeyValuePair{})
	configureTypes(database, types.User{}, types.Settings{})
	configureTypes(database, types.DataPointMeta{}, types.Listener{}, types.Emitter{})

	DB = database
}

func InitContent() {
}

func configureTypes(database *gorm.DB, datatypes ...interface{}) {
	for _, datatype := range datatypes {
		stmt := &gorm.Statement{DB: database}
		stmt.Parse(datatype)
		name := stmt.Schema.Table
		types.RegisterType(name, datatype)
		database.AutoMigrate(datatype)
	}
}
