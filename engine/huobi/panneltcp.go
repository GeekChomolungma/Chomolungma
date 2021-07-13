package huobi

import (
	"net"

	"github.com/GeekChomolungma/Chomolungma/config"
	"github.com/GeekChomolungma/Chomolungma/logging/applogger"
)

func subscribeMarketInfoTcp() {
	conn, err := net.Dial("tcp", config.GatewaySetting.GatewayTcpHost)
	if err != nil {
		applogger.Error("subscribeMarketInfoTcp dial failed:", err.Error())
	}
	conn.Write([]byte("hello"))
}

func readLoop(net.Conn) {}
