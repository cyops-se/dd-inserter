package main

import (
	"container/list"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	_ "github.com/lib/pq"
)

type DataPoint struct {
	Time    time.Time   `json:"t"`
	Name    string      `json:"n"`
	Value   interface{} `json:"v"`
	Quality int         `json:"q"`
}

type DataMessage struct {
	Counter uint64      `json:"counter"`
	Count   int         `json:"count"`
	Points  []DataPoint `json:"points"`
}

type batch struct {
	builder strings.Builder
	count   uint64
}

var queueLock sync.Mutex
var queue *list.List

func main() {
	pghost := flag.String("host", "localhost", "The Postgres database server")
	pgport := flag.Int("port", 5432, "The Postgres database port")
	pguser := flag.String("user", "postgres", "Username that are granted access to the database")
	pgpassword := flag.String("password", "password", "User password")
	dbname := flag.String("db", "demo", "The Postgres database name")
	port := flag.Int("listen", 4357, "The inserter UDP listen port")
	authident := flag.Bool("ident", false, "Use local user identification, not password")
	flag.Parse()

	queue = list.New()

	psqlInfo := fmt.Sprintf("dbname=%s sslmode=disable", *dbname)
	if !*authident {
		psqlInfo = fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			*pghost, *pgport, *pguser, *pgpassword, *dbname)

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

	go processMessages(db)
	go ticker(db)
	go listener(db, *port)

	// Sleep forever
	<-(chan int)(nil)
}

func listener(db *sql.DB, port int) {
	p := make([]byte, 2048)
	addr := net.UDPAddr{
		Port: port,
		IP:   net.ParseIP("0.0.0.0"),
	}

	ser, err := net.ListenUDP("udp", &addr)
	if err != nil {
		fmt.Printf("Failed to listen %v\n", err)
		return
	}

	for {
		n, _, err := ser.ReadFromUDP(p)
		if err != nil {
			fmt.Printf("Some error  %v", err)
			continue
		}

		var msg DataMessage
		if err := json.Unmarshal(p[:n], &msg); err != nil {
			fmt.Println("Failed to unmarshal data, err:", err)
			return
		}

		queueLock.Lock()
		queue.PushBack(msg)
		queueLock.Unlock()
	}
}

func processMessages(db *sql.DB) {

	counter := uint64(0)
	b := initBatch()
	for {
		queueLock.Lock()
		e := queue.Front()
		queueLock.Unlock()

		if e == nil || e.Value == nil {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		msg := e.Value.(DataMessage)

		queueLock.Lock()
		queue.Remove(e)
		queueLock.Unlock()

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

func (b *batch) appendBatch(datapoints []DataPoint) bool {
	for _, v := range datapoints {
		switch v.Value.(type) {
		case time.Time: // Skip
		case string: // Skip
		case bool: // Skip
		default:
			if b.count > 0 {
				fmt.Fprintf(&b.builder, ",")
			}
			fmt.Fprintf(&b.builder, "('%s', '%s', %v, %d)", v.Time.Format(time.RFC3339), v.Name, v.Value, v.Quality)
			b.count++
		}
	}

	return b.count > 1000
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
