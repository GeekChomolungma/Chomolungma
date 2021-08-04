package handler

import (
	"net/http"

	"github.com/GeekChomolungma/Chomolungma/dtos"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

func placeOrder(c *gin.Context) {
	var Req dtos.HttpReqModel
	err := c.Bind(&Req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": dtos.CANNOT_PARSE_POST_BODY, "msg": "Sorry", "data": err.Error()})
		return
	}

	switch Req.AimSite {
	case "HuoBi":
		orderInfo := &dtos.OrderInfoReq{}
		err = jsoniter.UnmarshalFromString(Req.Body, orderInfo)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": dtos.CANNOT_PARSE_POST_BODY, "msg": "Sorry", "data": err.Error()})
			return
		}
		huobi.PlaceOrder(orderInfo.Symbol, orderInfo.Model, orderInfo.Amount, orderInfo.Price, orderInfo.Source)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": dtos.AIM_SITE_NOT_EXIST, "msg": "Sorry", "data": err.Error()})
	}
}
