package main

import (
	"github.com/GeekChomolungma/Chomolungma/logging/applogger"
	"github.com/gorilla/websocket"
)

var url string = "ws://65.52.174.232:9780/ws"

func main() {
	_, _, err1 := websocket.DefaultDialer.Dial(url, nil)
	if err1 != nil {
		applogger.Error("WebSocket connected error1: %s", err1.Error())
		return
	}
	_, _, err2 := websocket.DefaultDialer.Dial(url, nil)
	if err2 != nil {
		applogger.Error("WebSocket connected error2: %s", err2.Error())
		return
	}
}
