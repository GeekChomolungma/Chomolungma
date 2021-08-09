package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GeekChomolungma/Chomolungma/config"
	"github.com/GeekChomolungma/Chomolungma/engine"
	"github.com/GeekChomolungma/Chomolungma/handler"
	"github.com/GeekChomolungma/Chomolungma/logging/applogger"
)

func main() {
	fmt.Println("CHOMOLUNGMA!")

	// config server
	config.Setup()

	// setup the trade engine
	engine.EngineBus = &engine.TradeEngine{
		Cylinders: make(map[string]engine.Cylinder),
	}
	engine.EngineBus.Load()
	engine.EngineBus.Run()

	// start http server
	go handler.LocalServer()

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
