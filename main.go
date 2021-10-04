// Copyright 2021 cyops.se. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cyops-se/dd-inserter/db"
	"github.com/cyops-se/dd-inserter/emitters"
	"github.com/cyops-se/dd-inserter/engine"
	"github.com/cyops-se/dd-inserter/listeners"
	_ "github.com/lib/pq"
	"golang.org/x/sys/windows/svc"
)

func main() {
	// csvfile := flag.String("csv", "", "Filename of a CSV formatted file with timestamped data to import from in the format 'name,time,value,quality' (one line header, time in format 2019-01-01 00:06:00)")
	flag.Parse()

	isIntSess, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
	}
	if !isIntSess {
		// runService(svcName)
		return
	}

	db.ConnectDatabase()
	db.InitContent()

	listeners.Init()
	emitters.Init()
	engine.InitDispatchers()

	go RunWeb()
	go emitters.RunDispatch()

	// Sleep until interrupted
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Exiting (waiting 1 sec) ...")
	time.Sleep(time.Second * 1)
}
