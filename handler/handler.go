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
	r.POST("/api/v1/common/timestamp", getTimestamp)

	// for account info
	r.POST("/api/v1/account/accountinfo", getAccountInfo)
	r.POST("/api/v1/account/accountbalance", getAccountBalance)

	// order action
	r.POST("/api/v1/order/placeorder", placeOrder)
	r.POST("/api/v1/order/cancelorder", cancelOrderById)

	// server run!
	r.Run(config.ServerSetting.Host)
}
