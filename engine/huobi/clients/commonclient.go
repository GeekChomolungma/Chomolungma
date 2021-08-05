package clients

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/GeekChomolungma/Chomolungma/dtos"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/internal"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/internal/requestbuilder"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/model"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/model/common"
)

// Responsible to get common information
type CommonClient struct {
	gatewayHost      string
	publicUrlBuilder *requestbuilder.PublicUrlBuilder
}

// Initializer
func (p *CommonClient) Init(gatewayHost string, host string) *CommonClient {
	p.gatewayHost = gatewayHost
	p.publicUrlBuilder = new(requestbuilder.PublicUrlBuilder).Init(host)
	return p
}

func (p *CommonClient) BuildAndPostGatewayUrl(request *dtos.BaseReqModel, originUrl string) (*dtos.BaseRspModel, error) {
	urlMsg := p.publicUrlBuilder.Build(originUrl, nil)
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

func (p *CommonClient) GetSystemStatus() (string, error) {
	url := "https://status.huobigroup.com/api/v2/summary.json"
	getResp, getErr := internal.HttpGet(url)
	if getErr != nil {
		return "", getErr
	}

	return getResp, nil
}

// Returns current market status
func (p *CommonClient) GetMarketStatus() (*common.MarketStatus, error) {
	url := p.publicUrlBuilder.Build("/v2/market-status", nil)
	getResp, getErr := internal.HttpGet(url)
	if getErr != nil {
		return nil, getErr
	}

	result := common.GetMarketStatusResponse{}
	jsonErr := json.Unmarshal([]byte(getResp), &result)
	if jsonErr != nil {
		return nil, jsonErr
	}
	if result.Code == 200 && &result.Data != nil {
		return &result.Data, nil
	}
	return nil, errors.New(getResp)
}

// Get all Supported Trading Symbol
// This endpoint returns all Huobi's supported trading symbol.
func (p *CommonClient) GetSymbols() ([]common.Symbol, error) {
	// create post body to gateway
	request := &dtos.BaseReqModel{
		AimSite: "HuoBi",
		Method:  "GET",
	}

	rawRsp, err := p.BuildAndPostGatewayUrl(request, "/v1/common/symbols")
	if err != nil {
		return nil, err
	}

	result := common.GetSymbolsResponse{}
	jsonErr := json.Unmarshal([]byte(rawRsp.Data), &result)
	if jsonErr != nil {
		return nil, jsonErr
	}
	if result.Status == "ok" && result.Data != nil {
		return result.Data, nil
	}
	return nil, errors.New(rawRsp.Data)
}

// Get all Supported Currencies
// This endpoint returns all Huobi's supported trading currencies.
func (p *CommonClient) GetCurrencys() ([]string, error) {
	url := p.publicUrlBuilder.Build("/v1/common/currencys", nil)
	getResp, getErr := internal.HttpGet(url)
	if getErr != nil {
		return nil, getErr
	}

	result := common.GetCurrenciesResponse{}
	jsonErr := json.Unmarshal([]byte(getResp), &result)

	if jsonErr != nil {
		return nil, jsonErr
	}
	if result.Status == "ok" && result.Data != nil {
		return result.Data, nil
	}
	return nil, errors.New(getResp)
}

// APIv2 - Currency & Chains
// API user could query static reference information for each currency, as well as its corresponding chain(s). (Public Endpoint)
func (p *CommonClient) GetV2ReferenceCurrencies(optionalRequest common.GetV2ReferenceCurrencies) ([]common.CurrencyChain, error) {
	request := new(model.GetRequest).Init()
	if optionalRequest.Currency != "" {
		request.AddParam("currency", optionalRequest.Currency)
	}
	if optionalRequest.AuthorizedUser != "" {
		request.AddParam("authorizedUser", optionalRequest.AuthorizedUser)
	}

	url := p.publicUrlBuilder.Build("/v2/reference/currencies", request)

	getResp, getErr := internal.HttpGet(url)
	if getErr != nil {
		return nil, getErr
	}

	result := common.GetV2ReferenceCurrenciesResponse{}
	jsonErr := json.Unmarshal([]byte(getResp), &result)

	if jsonErr != nil {
		return nil, jsonErr
	}

	if result.Code == 200 && result.Data != nil {
		return result.Data, nil
	}

	return nil, errors.New(result.Message)
}

// Get Current Timestamp
// This endpoint returns the current timestamp, i.e. the number of milliseconds that have elapsed since 00:00:00 UTC on 1 January 1970.
func (p *CommonClient) GetTimestamp() (int, error) {
	request := &dtos.BaseReqModel{
		AimSite: "HuoBi",
		Method:  "GET",
	}

	rawRsp, err := p.BuildAndPostGatewayUrl(request, "/v1/common/timestamp")
	if err != nil {
		return 0, err
	}

	result := common.GetTimestampResponse{}
	jsonErr := json.Unmarshal([]byte(rawRsp.Data), &result)

	if jsonErr != nil {
		return 0, jsonErr
	}
	if result.Status == "ok" && result.Data != 0 {
		return result.Data, nil
	}
	return 0, errors.New(rawRsp.Data)
}
