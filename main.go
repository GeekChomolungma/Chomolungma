package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "net/http/pprof"

	"github.com/GeekChomolungma/Chomolungma/config"
	"github.com/GeekChomolungma/Chomolungma/db/mongoInc"
	"github.com/GeekChomolungma/Chomolungma/engine"
	"github.com/GeekChomolungma/Chomolungma/handler"
	"github.com/GeekChomolungma/Chomolungma/logging/applogger"
)

func main() {
	fmt.Println("CHOMOLUNGMA!")

	// config server
	config.Setup("./my.ini")

	// setup the trade engine
	engine.EngineBus = &engine.TradeEngine{
		Cylinders: make(map[string]engine.Cylinder),
	}

	mongoInc.Init(config.MongoSetting.Uri)
	engine.EngineBus.Load()
	engine.EngineBus.Run()

	// start http server
	go handler.LocalServer()

	go http.ListenAndServe(":6060", nil)

	c := make(chan os.Signal, 5)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGABRT, syscall.SIGTERM)
	for {
		select {
		case sig := <-c:
			applogger.Info("Capture a System Call: %s", sig.String())
			engine.EngineBus.Stop()
			return
		default:
			time.Sleep(time.Duration(3) * time.Second)
		}
	}
}
