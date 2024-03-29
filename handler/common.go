package handler

import (
	"net/http"

	"github.com/GeekChomolungma/Chomolungma/config"
	"github.com/GeekChomolungma/Chomolungma/dtos"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi"
	"github.com/gin-gonic/gin"
)

func getSymbols(c *gin.Context) {
	var Req dtos.HttpReqModel
	err := c.Bind(&Req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": dtos.CANNOT_PARSE_POST_BODY, "msg": "Sorry", "data": err.Error()})
		return
	}

	switch Req.AimSite {
	case "HuoBi":
		symbols, err := huobi.GetSymbols()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": dtos.CANNOT_GET_SYMBOLS, "msg": "Sorry", "data": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": dtos.OK, "msg": "OK", "data": symbols})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": dtos.AIM_SITE_NOT_EXIST, "msg": "Sorry", "data": err.Error()})
	}
}

func getTimestamp(c *gin.Context) {
	var Req dtos.HttpReqModel
	err := c.Bind(&Req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": dtos.CANNOT_PARSE_POST_BODY, "msg": "Sorry", "data": err.Error()})
		return
	}

	switch Req.AimSite {
	case "HuoBi":
		timestamp, err := huobi.GetTimestamp()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": dtos.CANNOT_GET_SYMBOLS, "msg": "Sorry", "data": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"code": dtos.OK, "msg": "OK", "data": timestamp})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": dtos.AIM_SITE_NOT_EXIST, "msg": "Sorry", "data": err.Error()})
	}
}

func reloadKeys(c *gin.Context) {
	ok := config.ReloadKeys()
	if ok {
		c.JSON(http.StatusOK, gin.H{"code": dtos.OK, "msg": "OK", "data": "reloaded keys."})
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"code": dtos.CANNOT_RELOAD_KEYS, "msg": "Sorry", "data": ""})
	}
}

func ticksValidation(c *gin.Context) {
	var Req dtos.DBValidationReq
	err := c.Bind(&Req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": dtos.CANNOT_PARSE_POST_BODY, "msg": "Sorry", "data": err.Error()})
		return
	}

	if Req.Secret != "Chomolungma" {
		c.JSON(http.StatusBadRequest, gin.H{"code": dtos.AIM_SITE_POST_ERROR, "msg": "Sorry", "data": "Secret error."})
		return
	}

	switch Req.AimSite {
	case "HuoBi":
		huobi.TicksValidation(Req.Collection)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": dtos.AIM_SITE_NOT_EXIST, "msg": "Sorry", "data": err.Error()})
	}
}
