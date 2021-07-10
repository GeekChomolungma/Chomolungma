package main

import (
	"fmt"

	"github.com/GeekChomolungma/Chomolungma/config"
	"github.com/GeekChomolungma/Chomolungma/engine"
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

	// // register gin server
	// r := gin.Default()
	// // for user login or signup
	// r.POST("/api/v1/account/action", handler.Action)
	// // server run!
	// r.Run(config.ServerSetting.Host)
}
