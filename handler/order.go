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
		orderInfo := &dtos.OrderPlace{}
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

func cancelOrderById(c *gin.Context) {
	var Req dtos.HttpReqModel
	err := c.Bind(&Req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": dtos.CANNOT_PARSE_POST_BODY, "msg": "Sorry", "data": err.Error()})
		return
	}

	switch Req.AimSite {
	case "HuoBi":
		orderCancel := &dtos.OrderCancel{}
		err = jsoniter.UnmarshalFromString(Req.Body, orderCancel)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": dtos.CANNOT_PARSE_POST_BODY, "msg": "Sorry", "data": err.Error()})
			return
		}
		huobi.CancelOrderById(orderCancel.OrderID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": dtos.AIM_SITE_NOT_EXIST, "msg": "Sorry", "data": err.Error()})
	}
}

func getOrderById(c *gin.Context) {
	var Req dtos.HttpReqModel
	err := c.Bind(&Req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": dtos.CANNOT_PARSE_POST_BODY, "msg": "Sorry", "data": err.Error()})
		return
	}

	switch Req.AimSite {
	case "HuoBi":
		orderQuery := &dtos.OrderQuery{}
		err = jsoniter.UnmarshalFromString(Req.Body, orderQuery)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": dtos.CANNOT_PARSE_POST_BODY, "msg": "Sorry", "data": err.Error()})
			return
		}
		huobi.GetOrderById(orderQuery.OrderID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": dtos.AIM_SITE_NOT_EXIST, "msg": "Sorry", "data": err.Error()})
	}
}
