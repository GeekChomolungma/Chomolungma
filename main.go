package main

import (
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/GeekChomolungma/Chomolungma/config"
	"github.com/GeekChomolungma/Chomolungma/engine"
)

func main() {
	fmt.Println("CHOMOLUNGMA!")

	// config server
	config.Setup()

	// setup mongodb

	// setup the trade engine
	engine.EngineBus = &engine.TradeEngine{
		Cylinders: make(map[string]engine.Cylinder),
	}
	engine.EngineBus.Load()
	engine.EngineBus.Run()

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

	// // register gin server
	// r := gin.Default()
	// // for user login or signup
	// r.POST("/api/v1/account/action", handler.Action)
	// // server run!
	// r.Run(config.ServerSetting.Host)
}
