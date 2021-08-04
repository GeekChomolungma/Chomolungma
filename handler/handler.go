package handler

import (
	"github.com/GeekChomolungma/Chomolungma/config"
	"github.com/gin-gonic/gin"
)

// handler is used for strategy
// This is a warpper for general trade stuff

func LocalServer() {
	// register gin server
	r := gin.Default()

	// common status info
	r.POST("/api/v1/common/symbols", getSymbols)

	// for account info
	r.POST("/api/v1/account/accountinfo", getAccountInfo)
	r.POST("/api/v1/account/accountbalance", getAccountBalance)

	// order action
	r.POST("/api/v1/order/placeorder", placeOrder)

	// server run!
	r.Run(config.ServerSetting.Host)
}
