package handler

import (
	"net/http"

	"github.com/GeekChomolungma/Chomolungma/dtos"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi"
	"github.com/gin-gonic/gin"
	jsoniter "github.com/json-iterator/go"
)

func getAccountInfo(c *gin.Context) {
	var Req dtos.HttpReqModel
	err := c.Bind(&Req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": dtos.CANNOT_PARSE_POST_BODY, "msg": "Sorry", "data": err.Error()})
		return
	}

	switch Req.AimSite {
	case "HuoBi":
		huobi.GetAccountInfo(Req.AccountID)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": dtos.AIM_SITE_NOT_EXIST, "msg": "Sorry", "data": err.Error()})
	}
}

func getAccountBalance(c *gin.Context) {
	var Req dtos.HttpReqModel
	err := c.Bind(&Req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": dtos.CANNOT_PARSE_POST_BODY, "msg": "Sorry", "data": err.Error()})
		return
	}

	switch Req.AimSite {
	case "HuoBi":
		// get all currency
		balance, err := huobi.GetAccountBalance(Req.AccountID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": dtos.CANNOT_GET_ACCOUNT_BALANCE, "msg": "Sorry", "data": err.Error()})
			return
		}

		// get currency type
		balanceReq := &dtos.CurrencyBalanceReq{}
		err = jsoniter.UnmarshalFromString(Req.Body, balanceReq)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"code": dtos.CANNOT_PARSE_POST_BODY, "msg": "Sorry", "data": err.Error()})
			return
		}

		// get balance
		for _, v := range balance.List {
			if v.Currency == balanceReq.Currency {
				if v.Type == "trade" {
					c.JSON(http.StatusOK, gin.H{"code": dtos.OK, "msg": "OK", "data": v.Balance})
					return
				}
			}
		}

		c.JSON(http.StatusBadRequest, gin.H{"code": dtos.CANNOT_GET_ACCOUNT_BALANCE, "msg": "OK", "data": ""})
	default:
		c.JSON(http.StatusBadRequest, gin.H{"code": dtos.AIM_SITE_NOT_EXIST, "msg": "Sorry", "data": err.Error()})
	}
}
