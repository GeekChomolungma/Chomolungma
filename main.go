package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/GeekChomolungma/Chomolungma/config"
	"github.com/GeekChomolungma/Chomolungma/engine"
	"github.com/GeekChomolungma/Chomolungma/handler"
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
	signal.Notify(c)
	for {
		select {
		case <-c:
			engine.EngineBus.Stop()
			return
		default:
			time.Sleep(time.Duration(3) * time.Second)
		}
	}
}
