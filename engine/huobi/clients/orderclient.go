package clients

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/GeekChomolungma/Chomolungma/dtos"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/internal"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/internal/requestbuilder"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/model"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/model/order"
	jsoniter "github.com/json-iterator/go"
)

// Responsible to operate on order
type OrderClient struct {
	gatewayHost       string
	privateUrlBuilder *requestbuilder.PrivateUrlBuilder
}

// Initializer
func (p *OrderClient) Init(gatewayHost string, accessKey string, secretKey string, host string) *OrderClient {
	p.gatewayHost = gatewayHost
	p.privateUrlBuilder = new(requestbuilder.PrivateUrlBuilder).Init(accessKey, secretKey, host)
	return p
}

func (p *OrderClient) BuildAndPostGatewayUrl(request *dtos.BaseReqModel, originUrl string) (*dtos.BaseRspModel, error) {
	urlMsg := p.privateUrlBuilder.Build(request.Method, originUrl, nil)
	request.Url = urlMsg
	postBody, jsonErr := model.ToJson(request)
	if jsonErr != nil {
		return nil, jsonErr
	}

	// build url to gate way
	url := fmt.Sprintf("http://%s/api/v1/Chomolungma/entrypoint", p.gatewayHost)
	gatewayRsp, postErr := internal.HttpPost(url, postBody)
	if postErr != nil {
		return nil, postErr
	}

	// first parse the gin rsp
	rawRsp := &dtos.BaseRspModel{}
	jsonErr = json.Unmarshal([]byte(gatewayRsp), rawRsp)
	if jsonErr != nil {
		return nil, jsonErr
	}

	// then parse the data in gin rsp
	if rawRsp.Code != dtos.OK {
		return nil, errors.New("ERROR: Gateway response a error msg")
	}

	return rawRsp, nil
}

// Place a new order and send to the exchange to be matched.
func (p *OrderClient) PlaceOrder(request *order.PlaceOrderRequest) (*order.PlaceOrderResponse, error) {
	postBodyOrigin, jsonErr := model.ToJson(request)
	if jsonErr != nil {
		return nil, jsonErr
	}

	// create post body to gateway
	requestGateway := &dtos.BaseReqModel{
		AimSite: "HuoBi",
		Method:  "POST",
		Body:    postBodyOrigin,
	}

	// build gateway url and post it
	rawRsp, err := p.BuildAndPostGatewayUrl(requestGateway, "/v1/order/orders/place")
	if err != nil {
		return nil, err
	}

	result := &order.PlaceOrderResponse{}
	jsonErr = jsoniter.Unmarshal([]byte(rawRsp.Data), result)
	if jsonErr != nil {
		return nil, jsonErr
	}
	if result.Status == "ok" {
		return result, nil
	}

	return nil, errors.New(rawRsp.Data)
}

// Place multipler orders (at most 10 orders)
func (p *OrderClient) PlaceOrders(request []order.PlaceOrderRequest) (*order.PlaceOrdersResponse, error) {

	postBody, jsonErr := model.ToJson(request)

	url := p.privateUrlBuilder.Build("POST", "/v1/order/batch-orders", nil)
	postResp, postErr := internal.HttpPost(url, string(postBody))
	if postErr != nil {
		return nil, postErr
	}

	result := order.PlaceOrdersResponse{}
	jsonErr = json.Unmarshal([]byte(postResp), &result)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return &result, nil
}

// Cancel an order by order id
func (p *OrderClient) CancelOrderById(orderId string) (*order.CancelOrderByIdResponse, error) {
	path := fmt.Sprintf("/v1/order/orders/%s/submitcancel", orderId)
	url := p.privateUrlBuilder.Build("POST", path, nil)
	postResp, postErr := internal.HttpPost(url, "")
	if postErr != nil {
		return nil, postErr
	}

	result := order.CancelOrderByIdResponse{}
	jsonErr := json.Unmarshal([]byte(postResp), &result)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return &result, nil
}

// Cancel an order by client order id
func (p *OrderClient) CancelOrderByClientOrderId(clientOrderId string) (*order.CancelOrderByClientResponse, error) {
	url := p.privateUrlBuilder.Build("POST", "/v1/order/orders/submitCancelClientOrder", nil)
	body := fmt.Sprintf("{\"client-order-id\":\"%s\"}", clientOrderId)
	postResp, postErr := internal.HttpPost(url, body)
	if postErr != nil {
		return nil, postErr
	}

	result := order.CancelOrderByClientResponse{}
	jsonErr := json.Unmarshal([]byte(postResp), &result)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return &result, nil
}

// Returns all open orders which have not been filled completely.
func (p *OrderClient) GetOpenOrders(request *model.GetRequest) (*order.GetOpenOrdersResponse, error) {
	url := p.privateUrlBuilder.Build("GET", "/v1/order/openOrders", request)
	getResp, getErr := internal.HttpGet(url)
	if getErr != nil {
		return nil, getErr
	}

	result := order.GetOpenOrdersResponse{}
	jsonErr := json.Unmarshal([]byte(getResp), &result)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return &result, nil
}

// Submit cancellation for multiple orders at once with given criteria.
func (p *OrderClient) CancelOrdersByCriteria(request *order.CancelOrdersByCriteriaRequest) (*order.CancelOrdersByCriteriaResponse, error) {
	postBody, jsonErr := model.ToJson(request)

	url := p.privateUrlBuilder.Build("POST", "/v1/order/orders/batchCancelOpenOrders", nil)
	postResp, postErr := internal.HttpPost(url, string(postBody))
	if postErr != nil {
		return nil, postErr
	}

	result := order.CancelOrdersByCriteriaResponse{}
	jsonErr = json.Unmarshal([]byte(postResp), &result)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return &result, nil
}

// Submit cancellation for multiple orders at once with given ids
func (p *OrderClient) CancelOrdersByIds(request *order.CancelOrdersByIdsRequest) (*order.CancelOrdersByIdsResponse, error) {
	postBody, jsonErr := model.ToJson(request)

	url := p.privateUrlBuilder.Build("POST", "/v1/order/orders/batchcancel", nil)
	postResp, postErr := internal.HttpPost(url, string(postBody))
	if postErr != nil {
		return nil, postErr
	}

	result := order.CancelOrdersByIdsResponse{}
	jsonErr = json.Unmarshal([]byte(postResp), &result)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return &result, nil
}

// Returns the detail of one order by order id
func (p *OrderClient) GetOrderById(orderId string) (*order.GetOrderResponse, error) {
	path := fmt.Sprintf("/v1/order/orders/%s", orderId)
	url := p.privateUrlBuilder.Build("GET", path, nil)
	getResp, getErr := internal.HttpGet(url)
	if getErr != nil {
		return nil, getErr
	}

	result := order.GetOrderResponse{}
	jsonErr := json.Unmarshal([]byte(getResp), &result)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return &result, nil
}

// Returns the detail of one order by client order id
func (p *OrderClient) GetOrderByCriteria(request *model.GetRequest) (*order.GetOrderResponse, error) {
	url := p.privateUrlBuilder.Build("GET", "/v1/order/orders/getClientOrder", request)
	getResp, getErr := internal.HttpGet(url)
	if getErr != nil {
		return nil, getErr
	}

	result := order.GetOrderResponse{}
	jsonErr := json.Unmarshal([]byte(getResp), &result)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return &result, nil
}

// Returns the match result of an order.
func (p *OrderClient) GetMatchResultsById(orderId string) (*order.GetMatchResultsResponse, error) {
	path := fmt.Sprintf("/v1/order/orders/%s/matchresults", orderId)
	url := p.privateUrlBuilder.Build("GET", path, nil)
	getResp, getErr := internal.HttpGet(url)
	if getErr != nil {
		return nil, getErr
	}

	result := order.GetMatchResultsResponse{}
	jsonErr := json.Unmarshal([]byte(getResp), &result)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return &result, nil
}

// Returns orders based on a specific searching criteria.
func (p *OrderClient) GetHistoryOrders(request *model.GetRequest) (*order.GetHistoryOrdersResponse, error) {
	url := p.privateUrlBuilder.Build("GET", "/v1/order/orders", request)
	getResp, getErr := internal.HttpGet(url)
	if getErr != nil {
		return nil, getErr
	}

	result := order.GetHistoryOrdersResponse{}
	jsonErr := json.Unmarshal([]byte(getResp), &result)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return &result, nil
}

// Returns orders based on a specific searching criteria (within 48 hours)
func (p *OrderClient) GetLast48hOrders(request *model.GetRequest) (*order.GetHistoryOrdersResponse, error) {
	url := p.privateUrlBuilder.Build("GET", "/v1/order/history", request)
	getResp, getErr := internal.HttpGet(url)
	if getErr != nil {
		return nil, getErr
	}

	result := order.GetHistoryOrdersResponse{}
	jsonErr := json.Unmarshal([]byte(getResp), &result)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return &result, nil
}

// Returns the match results of past and open orders based on specific search criteria.
func (p *OrderClient) GetMatchResultsByCriteria(request *model.GetRequest) (*order.GetMatchResultsResponse, error) {
	url := p.privateUrlBuilder.Build("GET", "/v1/order/matchresults", request)
	getResp, getErr := internal.HttpGet(url)
	if getErr != nil {
		return nil, getErr
	}

	result := order.GetMatchResultsResponse{}
	jsonErr := json.Unmarshal([]byte(getResp), &result)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return &result, nil
}

// Returns the current transaction fee rate applied to the user.
func (p *OrderClient) GetTransactFeeRate(request *model.GetRequest) (*order.GetTransactFeeRateResponse, error) {
	url := p.privateUrlBuilder.Build("GET", "/v2/reference/transact-fee-rate", request)
	getResp, getErr := internal.HttpGet(url)
	if getErr != nil {
		return nil, getErr
	}

	result := order.GetTransactFeeRateResponse{}
	jsonErr := json.Unmarshal([]byte(getResp), &result)
	if jsonErr != nil {
		return nil, jsonErr
	}

	return &result, nil
}
