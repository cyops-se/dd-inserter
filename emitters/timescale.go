package emitters

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cyops-se/dd-inserter/types"
	"github.com/lib/pq"
)

type TimescaleEmitter struct {
	Host        string                `json:"host"`
	Port        int                   `json:"port"`
	User        string                `json:"user"`
	Password    string                `json:"password"`
	Authident   bool                  `json:"authident"`
	Database    string                `json:"database"`
	Batchsize   int                   `json:"batchsize"`
	err         error                 `json:"-" gorm:"-"`
	initialized bool                  `json:"-" gorm:"-"`
	messages    chan *types.DataPoint `json:"-" gorm:"-"`
	builder     strings.Builder       `json:"-" gorm:"-"`
	count       uint64                `json:"-" gorm:"-"`
}

var debug *bool
var batchSize *int
var TimescaleDB *sql.DB
var ids map[string]int

func (emitter *TimescaleEmitter) InitEmitter() error {
	if err := emitter.connectdb(); err == nil {
		emitter.initialized = true
	}

	ids = make(map[string]int)
	emitter.initBatch()

	emitter.messages = make(chan *types.DataPoint, 2000)
	go emitter.processMessages()

	return emitter.err
}

func (emitter *TimescaleEmitter) ProcessMessage(dp *types.DataPoint) {
	if emitter.initialized == false {
		return
	}

	emitter.messages <- dp
}

func (emitter *TimescaleEmitter) processMessages() {

	for {
		dp := <-emitter.messages

		var err error
		_, isfloat64 := dp.Value.(float64)
		_, isint := dp.Value.(int)
		_, isuint64 := dp.Value.(uint64)
		if !isfloat64 && !isint && !isuint64 {
			log.Println("datapoint not float64 or int or uint64:", dp.Name)
			return
		}

		// use 'ids' as a local datapoint name cache to resolve id
		// if not in the cache, get it from the database
		// if not in the database, insert a new meta item and get the new id
		id, ok := ids[dp.Name]
		if !ok {
			if err = TimescaleDB.QueryRow("select tag_id from measurements.tags where name=$1", dp.Name).Scan(&id); err != nil {
				if err == sql.ErrNoRows {
					TimescaleDB.QueryRow("insert into measurements.tags (name) values ($1) returning tag_id", dp.Name).Scan(&id)
				}
			}
			ids[dp.Name] = id
		}

		if err == nil {
			if emitter.appendPoint(id, dp) {
				emitter.insertBatch()
				emitter.initBatch()
			}
		} else {
			fmt.Println("TIMESCALEDB insert process data failed, err:", err.Error())
		}
	}
}

func (emitter *TimescaleEmitter) ProcessMeta(dp *types.DataPointMeta) {
	// fmt.Println("TIMESCALE emitter processing META message")

	var id int
	if emitter.rowExists("select name from measurements.tags where name=$1", dp.Name) == false {
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

	log.Printf("TIMESCALE server connected: %s", emitter.Host)
	return emitter.err
}

func (emitter *TimescaleEmitter) rowExists(query string, args ...interface{}) bool {
	var exists bool
	query = fmt.Sprintf("SELECT exists (%s)", query)
	err := TimescaleDB.QueryRow(query, args...).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		log.Fatalf("error checking if row exists '%s' %v", args, err)
	}
	return exists
}

func (emitter *TimescaleEmitter) initBatch() {
	emitter.count = 0
	emitter.builder.Reset()
	emitter.builder.Grow(4096)

	fmt.Fprintf(&emitter.builder, "insert into measurements.raw_measurements (tag, time, value, quality) values ")
}

func (emitter *TimescaleEmitter) appendPoint(id int, v *types.DataPoint) bool {
	switch v.Value.(type) {
	case time.Time: // Skip
	case string: // Skip
	case bool:
		if v.Value != nil {
			if emitter.count > 0 {
				fmt.Fprintf(&emitter.builder, ",")
			}
			value := 0.0
			if v.Value.(bool) {
				value = 1.0
			}

			fmt.Fprintf(&emitter.builder, "(%d, '%s', %v, %d)", id, v.Time.Format(time.RFC3339Nano), value, v.Quality)
			emitter.count++
		}
	default:
		if v.Value != nil {
			if emitter.count > 0 {
				fmt.Fprintf(&emitter.builder, ",")
			}

			// log.Printf("time %s, %v", v.Time.Format(time.RFC3339Nano), v.Time)
			fmt.Fprintf(&emitter.builder, "(%d, '%s', %v, %d)", id, v.Time.Format(time.RFC3339Nano), v.Value, v.Quality)
			emitter.count++
		}
	}

	return emitter.count > uint64(emitter.Batchsize)
}

func (emitter *TimescaleEmitter) insertBatch() error {
	if emitter.count > 0 {
		fmt.Fprintf(&emitter.builder, ";")
		insert := emitter.builder.String()
		_, err := TimescaleDB.Exec(insert)
		if err != nil {
			switch err.(type) {
			default:
				log.Printf("failed to insert: %#v", err)
				return err
			case *pq.Error:
				if err.(*pq.Error).Code != "23505" { // duplicate key
					log.Printf("failed to insert: %#v", err)
					fmt.Println(insert)
					emitter.initBatch()
					return err
				}
			}
		}

		// log.Printf("%d rows inserted to %s, database %s, result %#v", emitter.count, emitter.Host, emitter.Database, result)
		// log.Println(insert)
	}

	return nil
}
