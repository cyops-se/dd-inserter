package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"net"
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
	Count  int         `json:"count"`
	Points []DataPoint `json:"points"`
}

func main() {
	pghost := flag.String("host", "localhost", "The Postgres database server")
	pgport := flag.Int("port", 5432, "The Postgres database port")
	pguser := flag.String("user", "postgres", "Username that are granted access to the database")
	pgpassword := flag.String("password", "password", "User password")
	dbname := flag.String("db", "demo", "The Postgres database name")
	port := flag.Int("listen", 4357, "The inserter UDP listen port")
	flag.Parse()

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		*pghost, *pgport, *pguser, *pgpassword, *dbname)

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

	go ticker(db)
	go listener(db, *port)

	// Sleep forever
	<-(chan int)(nil)
}

func insert(db *sql.DB, t time.Time, n string, v interface{}, q int) error {
	statement := `insert into measurements(time, name, value, quality) values ($1, $2, $3, $4)`
	if _, err := db.Exec(statement, t, n, v, q); err != nil {
		// fmt.Println("Failed to insert data into the database, err:", err)
		return err
	}

	return nil
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
		// n, remoteaddr, err := ser.ReadFromUDP(p)
		// fmt.Printf("Read a message from %v %s \n", remoteaddr, p)
		if err != nil {
			fmt.Printf("Some error  %v", err)
			continue
		}

		var msg DataMessage
		if err = json.Unmarshal(p[:n], &msg); err != nil {
			fmt.Println("Failed to unmarshal data, err:", err)
			return
		}

		count := 0
		if msg.Count > 0 {
			for _, v := range msg.Points {
				if err = insert(db, v.Time, v.Name, v.Value, v.Quality); err == nil {
					count++
				}
			}

			fmt.Println(count, "of", len(msg.Points), "items inserted")
		}
	}
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

		insert(db, time.Now(), "testpoint.1", value, 1.0)
		insert(db, time.Now(), "testpoint.2", 50.0-value, 1.0)

		value += 0.1
		if value > 50.0 {
			value = 0.1
		}
	}
}
