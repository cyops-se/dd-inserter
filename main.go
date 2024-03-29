// Copyright 2021 cyops.se. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cyops-se/dd-inserter/db"
	"github.com/cyops-se/dd-inserter/emitters"
	"github.com/cyops-se/dd-inserter/engine"
	"github.com/cyops-se/dd-inserter/listeners"
	"github.com/cyops-se/dd-inserter/routes"
	"github.com/cyops-se/dd-inserter/types"
	"golang.org/x/sys/windows/svc"
)

var ctx types.Context
var GitVersion string
var GitCommit string

func main() {
	defer handlePanic()
	svcName := "dd-inserter"

	flag.StringVar(&ctx.Cmd, "cmd", "debug", "Windows service command (try 'usage' for more info)")
	flag.StringVar(&ctx.Wdir, "workdir", ".", "Specifies working directory for process (useful when running as service)")
	flag.BoolVar(&ctx.Trace, "trace", false, "Prints traces of OCP data to the console")
	flag.BoolVar(&ctx.Version, "v", false, "Prints the commit hash and exists")
	flag.Parse()

	routes.SysInfo.GitVersion = GitVersion
	routes.SysInfo.GitCommit = GitCommit

	if ctx.Version {
		fmt.Printf("dd-inserter version %s, commit: %s\n", routes.SysInfo.GitVersion, routes.SysInfo.GitCommit)
		return
	}

	if ctx.Cmd == "install" {
		if err := installService(svcName, "dd-inserter from cyops-se"); err != nil {
			log.Fatalf("failed to %s %s: %v", ctx.Cmd, svcName, err)
		}
		return
	} else if ctx.Cmd == "remove" {
		if err := removeService(svcName); err != nil {
			log.Fatalf("failed to %s %s: %v", ctx.Cmd, svcName, err)
		}
		return
	}

	inService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
	}
	if inService {
		runService(svcName, false)
		return
	}

	runService(svcName, true)
}

func runEngine() {
	defer handlePanic()

	db.ConnectDatabase(ctx)
	db.InitContent()

	listeners.RegisterType("NatsData", listeners.NATSDataListener{})
	listeners.RegisterType("NatsFile", listeners.NATSFileListener{})
	listeners.RegisterType("UdpData", listeners.UDPDataListener{})
	listeners.RegisterType("UdpMeta", listeners.UDPMetaListener{})
	listeners.RegisterType("UdpFile", listeners.UDPFileListener{})
	listeners.RegisterType("Cache", listeners.CacherListener{})
	// listeners.Init(ctx)
	listeners.LoadListeners(ctx)

	emitters.RegisterType("TimescaleDB", emitters.TimescaleEmitter{})
	emitters.RegisterType("RabbitMQ", emitters.RabbitMQEmitter{})
	emitters.LoadEmitters()

	engine.InitDispatchers()
	// engine.InitFileTransfer(ctx)
	engine.InitMonitor()

	go RunWeb()
	go emitters.RunDispatch()

	// Sleep until interrupted
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Exiting (waiting 1 sec) ...")
	time.Sleep(time.Second * 1)
}
