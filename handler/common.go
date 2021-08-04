package handler

import (
	"net/http"

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
