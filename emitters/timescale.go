package emitters

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/cyops-se/dd-inserter/types"
)

type TimescaleEmitter struct {
	EmitterBase
	Host      string `json:"host"`
	Port      int    `json:"port"`
	User      string `json:"username"`
	Password  string `json:"password"`
	Authident bool   `json:"authident"`
	Database  string `json:"database"`
	Batchsize int    `json:"batchsize"`
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

func (emitter *TimescaleEmitter) InitEmitter() {
	emitters = append(emitters, emitter)
}

func (emitter *TimescaleEmitter) ProcessMessage(dp *types.DataPoint) {
	// fmt.Println("TIMESCALE emitter processing message")
}

func (emitter *TimescaleEmitter) run(msg *types.DataMessage) {
	emitter.Host = "localhost"
	emitter.Port = 5432
	emitter.User = "admin"
	emitter.Password = "admin"
	emitter.Authident = false
	emitter.Database = "processdata"
	emitter.Batchsize = 1000

	psqlInfo := fmt.Sprintf("dbname=%s sslmode=disable", emitter.Database)
	if !emitter.Authident {
		psqlInfo = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			emitter.Host, emitter.Port, emitter.User, emitter.Password, emitter.Database)

	}

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		fmt.Println("Failed to connect to the database, err:", err)
		return
	}

	err = db.Ping()
	if err != nil {
		fmt.Println("Database PING failed, err:", err)
		return
	}

	defer db.Close()

	go insert(emitter, db)
	<-(chan int)(nil)
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
	case bool: // Skip
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
