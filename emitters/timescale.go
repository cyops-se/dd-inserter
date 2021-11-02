package emitters

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cyops-se/dd-inserter/engine"
	"github.com/cyops-se/dd-inserter/types"
	"github.com/lib/pq"
)

type TimescaleEmitter struct {
	Host        string `json:"host"`
	Port        int    `json:"port"`
	User        string `json:"user"`
	Password    string `json:"password"`
	Authident   bool   `json:"authident"`
	Database    string `json:"database"`
	Batchsize   int    `json:"batchsize"`
	err         error
	initialized bool
}

type generalParameters struct {
	debug bool
}

type receiveParameters struct {
	port int
}

type batch struct {
	builder strings.Builder
	count   uint64
}

var debug *bool
var batchSize *int
var TimescaleDB *sql.DB

func (emitter *TimescaleEmitter) InitEmitter() error {
	if err := emitter.connectdb(); err == nil {
		emitter.initialized = true
	}

	return emitter.err
}

// func (emitter *TimescaleEmitter) LoadSettingsJSON(settings string) error {
// 	// settings is a JSON object with all settings (serialized from TimescaleEmitter)
// 	return json.Unmarshal([]byte(settings), emitter)
// }

// func (emitter *TimescaleEmitter) GetSettingsJSON() (string, error) {
// 	settings, err := json.Marshal(emitter)
// 	if err != nil {
// 		db.Log("error", "Failed to save RabbitMQ settings", err.Error())
// 		return "", err
// 	}

// 	return string(settings), nil
// }

func (emitter *TimescaleEmitter) ProcessMessage(dp *types.DataPoint) {
	if emitter.initialized == false {
		return
	}

	var err error
	_, isfloat64 := dp.Value.(float64)
	_, isint := dp.Value.(int)
	_, isuint64 := dp.Value.(uint64)
	if !isfloat64 && !isint && !isuint64 {
		return
	}

	var id int
	// fmt.Println("TIMESCALE emitter processing message")
	if err = TimescaleDB.QueryRow("select tag_id from measurements.tags where name=$1", dp.Name).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			TimescaleDB.QueryRow("insert into measurements.tags (name) values ($1) returning tag_id", dp.Name).Scan(&id)
		}
	}

	if err == nil {
		if _, pqerr := TimescaleDB.Exec("insert into measurements.raw_measurements (tag, time, value, quality) values ($1, $2, $3, $4)", id, dp.Time, dp.Value, dp.Quality); pqerr != nil {
			if pqerr.(*pq.Error).Code != "23505" { // duplicate key
				fmt.Println("TIMESCALEDB process data insert failure:", dp.Time, id, dp.Name, err.Error())
			}
		}
	} else {
		fmt.Println("TIMESCALEDB insert process data failed, err:", err.Error())
	}
}

func (emitter *TimescaleEmitter) ProcessMeta(dp *types.DataPointMeta) {
	// fmt.Println("TIMESCALE emitter processing META message")

	var id int
	if rowExists("select name from measurements.tags where name=$1", dp.Name) == false {
		if err := TimescaleDB.QueryRow("insert into measurements.tags (name, description) values ($1, $2) returning tag_id", dp.Name, dp.Description).Scan(&id); err != nil {
			fmt.Println("TIMESCALE failed to insert,", err.Error())
		}
	} else {
		if _, err := TimescaleDB.Exec("update measurements.tags set description = $2 where name = $1", dp.Name, dp.Description); err != nil {
			fmt.Println("TIMESCALE failed to update,", err.Error())
		}
	}
}

func (emitter *TimescaleEmitter) GetStats() *types.EmitterStatistics {
	return nil
}

func (emitter *TimescaleEmitter) connectdb() error {
	psqlInfo := fmt.Sprintf("dbname=%s sslmode=disable", emitter.Database)
	if !emitter.Authident {
		psqlInfo = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			emitter.Host, emitter.Port, emitter.User, emitter.Password, emitter.Database)

	}

	TimescaleDB, emitter.err = sql.Open("postgres", psqlInfo)
	if emitter.err != nil {
		fmt.Println("Failed to connect to the database, err:", emitter.err)
		return emitter.err
	}

	emitter.err = TimescaleDB.Ping()
	if emitter.err != nil {
		fmt.Println("Database PING failed, err:", emitter.err)
		return emitter.err
	}

	// if _, emitter.err = TimescaleDB.Exec("insert into measurements.raw_measurements(time, tag, value, quality) values ('2021-09-10 13:00:00', 1, 45.6, 12)"); emitter.err != nil {
	// 	fmt.Println("TIMESCALE failed to insert,", emitter.err.Error())
	// }

	fmt.Println("TIMESCALE connected")
	return emitter.err
}

func rowExists(query string, args ...interface{}) bool {
	var exists bool
	query = fmt.Sprintf("SELECT exists (%s)", query)
	err := TimescaleDB.QueryRow(query, args...).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf("error checking if row exists '%s' %v", args, err)
	}
	return exists
}

func insert(ip *TimescaleEmitter, db *sql.DB) {

	go processMessages(db)
	go ticker(db)
}

func processMessages(db *sql.DB) {

	counter := uint64(0)
	b := initBatch()
	for {
		msg := &types.DataMessage{}

		if msg == nil {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		if msg.Counter-counter > 1 {
			fmt.Printf("MISSING MESSAGE, got counter: %d, expected %d, difference: %d\n", msg.Counter, counter+1, (msg.Counter - counter))
		}
		counter = msg.Counter

		if msg.Count > 0 {
			if full := b.appendBatch(msg.Points); full {
				b.insertBatch(db)
				b = initBatch()
				engine.NotifySubscribers("emitter.timescale", counter)
			}
		}
	}
}

func initBatch() *batch {
	b := &batch{}
	b.builder.Grow(4096)
	fmt.Fprintf(&b.builder, "insert into measurements(time, name, value, quality) values ")
	return b
}

func (b *batch) appendBatch(datapoints []types.DataPoint) bool {
	for _, v := range datapoints {
		b.appendPoint(&v)
	}

	return b.count > uint64(*batchSize)
}

func (b *batch) appendPoint(v *types.DataPoint) bool {
	switch v.Value.(type) {
	case time.Time: // Skip
	case string: // Skip
	case bool:
		if v.Value != nil {
			if b.count > 0 {
				fmt.Fprintf(&b.builder, ",")
			}
			value := 0.0
			if v.Value.(bool) {
				value = 1.0
			}

			fmt.Fprintf(&b.builder, "('%s', '%s', %v, %d)", v.Time.Format(time.RFC3339), v.Name, value, v.Quality)
			b.count++
		}
	default:
		if v.Value != nil {
			if b.count > 0 {
				fmt.Fprintf(&b.builder, ",")
			}
			fmt.Fprintf(&b.builder, "('%s', '%s', %v, %d)", v.Time.Format(time.RFC3339), v.Name, v.Value, v.Quality)
			b.count++
		}
	}

	return b.count > uint64(*batchSize)
}

func (b *batch) insertBatch(db *sql.DB) error {
	if b.count > 0 {
		fmt.Fprintf(&b.builder, ";")
		if _, err := db.Exec(b.builder.String()); err != nil {
			fmt.Println(b.builder.String())
			return err
		}
	}

	return nil
}

func ticker(db *sql.DB) {
	value := 0.1
	timer := time.NewTicker(1 * time.Second)
	for {
		<-timer.C

		if err := db.Ping(); err != nil {
			fmt.Println("Database PING failed, err:", err)
			return
		}

		value += 0.1
		if value > 50.0 {
			value = 0.1
		}
	}
}
