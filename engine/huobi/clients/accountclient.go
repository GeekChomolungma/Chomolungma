package clients

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/GeekChomolungma/Chomolungma/dtos"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/internal"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/internal/requestbuilder"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/model"
	"github.com/GeekChomolungma/Chomolungma/engine/huobi/model/account"
	jsoniter "github.com/json-iterator/go"
)

// Responsible to operate account
type AccountClient struct {
	gatewayHost       string
	privateUrlBuilder *requestbuilder.PrivateUrlBuilder
}

// Initializer
func (p *AccountClient) Init(gatewayHost string, accessKey string, secretKey string, host string) *AccountClient {
	p.gatewayHost = gatewayHost
	p.privateUrlBuilder = new(requestbuilder.PrivateUrlBuilder).Init(accessKey, secretKey, host)
	return p
}

func (p *AccountClient) BuildAndPostGatewayUrl(request *dtos.BaseReqModel, originUrl string) (*dtos.BaseRspModel, error) {
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

// Returns a list of accounts owned by this API user
func (p *AccountClient) GetAccountInfo() ([]account.AccountInfo, error) {
	// create post body to gateway
	request := &dtos.BaseReqModel{
		AimSite: "HuoBi",
		Method:  "GET",
	}

	// build gateway url and post it
	rawRsp, err := p.BuildAndPostGatewayUrl(request, "/v1/account/accounts")
	if err != nil {
		return nil, err
	}

	result := account.GetAccountInfoResponse{}
	jsonErr := jsoniter.Unmarshal([]byte(rawRsp.Data), &result)
	if jsonErr != nil {
		return nil, jsonErr
	}
	if result.Status == "ok" && result.Data != nil {
		return result.Data, nil
	}

	return nil, errors.New(rawRsp.Data)
}

// Returns the balance of an account specified by account id
func (p *AccountClient) GetAccountBalance(accountId string) (*account.AccountBalance, error) {
	// create post body to gateway
	request := &dtos.BaseReqModel{
		AimSite: "HuoBi",
		Method:  "GET",
	}

	// build gateway url and post it
	rawRsp, err := p.BuildAndPostGatewayUrl(request, "/v1/account/accounts/"+accountId+"/balance")
	if err != nil {
		return nil, err
	}

	result := account.GetAccountBalanceResponse{}
	jsonErr := jsoniter.Unmarshal([]byte(rawRsp.Data), &result)
	if jsonErr != nil {
		return nil, jsonErr
	}
	if result.Status == "ok" && result.Data != nil {
		return result.Data, nil
	}

	return nil, errors.New(rawRsp.Data)
}

// Returns the valuation of the total assets of the account in btc or fiat currency.
func (p *AccountClient) GetAccountAssetValuation(accountType string, valuationCurrency string, subUid int64) (*account.GetAccountAssetValuationResponse, error) {
	request := new(model.GetRequest).Init()
	request.AddParam("accountType", accountType)
	if valuationCurrency != "" {
		request.AddParam("valuationCurrency", valuationCurrency)
	}
	if subUid != 0 {
		request.AddParam("subUid", strconv.FormatInt(subUid, 10))
	}

	url := p.privateUrlBuilder.Build("GET", "/v2/account/asset-valuation", request)
	getResp, getErr := internal.HttpGet(url)
	if getErr != nil {
		return nil, getErr
	}
	result := account.GetAccountAssetValuationResponse{}
	jsonErr := json.Unmarshal([]byte(getResp), &result)
	if jsonErr != nil {
		return nil, jsonErr
	}
	if result.Code == 200 {
		return &result, nil
	}

	return nil, errors.New(getResp)
}

func (p *AccountClient) TransferAccount(request account.TransferAccountRequest) (*account.TransferAccountResponse, error) {
	postBody, jsonErr := model.ToJson(request)
	if jsonErr != nil {
		return nil, jsonErr
	}

	url := p.privateUrlBuilder.Build("POST", "/v1/account/transfer", nil)
	postResp, postErr := internal.HttpPost(url, postBody)
	if postErr != nil {
		return nil, postErr
	}

	result := account.TransferAccountResponse{}
	jsonErr = json.Unmarshal([]byte(postResp), &result)
	if jsonErr != nil {
		return nil, jsonErr
	}
	if result.Status != "ok" {
		return nil, errors.New(postResp)
	}

	return &result, nil
}

// Returns the amount changes of specified user's account
func (p *AccountClient) GetAccountHistory(accountId string, optionalRequest account.GetAccountHistoryOptionalRequest) ([]account.AccountHistory, error) {
	request := new(model.GetRequest).Init()
	request.AddParam("account-id", accountId)
	if optionalRequest.Currency != "" {
		request.AddParam("currency", optionalRequest.Currency)
	}
	if optionalRequest.Size != 0 {
		request.AddParam("size", strconv.Itoa(optionalRequest.Size))
	}
	if optionalRequest.EndTime != 0 {
		request.AddParam("end-time", strconv.FormatInt(optionalRequest.EndTime, 10))
	}
	if optionalRequest.Sort != "" {
		request.AddParam("sort", optionalRequest.Sort)
	}
	if optionalRequest.StartTime != 0 {
		request.AddParam("start-time", strconv.FormatInt(optionalRequest.StartTime, 10))
	}
	if optionalRequest.TransactTypes != "" {
		request.AddParam("transact-types", optionalRequest.TransactTypes)
	}

	url := p.privateUrlBuilder.Build("GET", "/v1/account/history", request)
	getResp, getErr := internal.HttpGet(url)
	if getErr != nil {
		return nil, getErr
	}
	result := account.GetAccountHistoryResponse{}
	jsonErr := json.Unmarshal([]byte(getResp), &result)
	if jsonErr != nil {
		return nil, jsonErr
	}
	if result.Status == "ok" && result.Data != nil {

		return result.Data, nil
	}

	return nil, errors.New(getResp)
}

// Returns the account ledger of specified user's account
func (p *AccountClient) GetAccountLedger(accountId string, optionalRequest account.GetAccountLedgerOptionalRequest) ([]account.Ledger, error) {
	request := new(model.GetRequest).Init()
	request.AddParam("accountId", accountId)
	if optionalRequest.Currency != "" {
		request.AddParam("currency", optionalRequest.Currency)
	}
	if optionalRequest.TransactTypes != "" {
		request.AddParam("transactTypes", optionalRequest.TransactTypes)
	}
	if optionalRequest.StartTime != 0 {
		request.AddParam("startTime", strconv.FormatInt(optionalRequest.StartTime, 10))
	}
	if optionalRequest.EndTime != 0 {
		request.AddParam("endTime", strconv.FormatInt(optionalRequest.EndTime, 10))
	}
	if optionalRequest.Sort != "" {
		request.AddParam("sort", optionalRequest.Sort)
	}
	if optionalRequest.Limit != 0 {
		request.AddParam("limit", strconv.Itoa(optionalRequest.Limit))
	}
	if optionalRequest.FromId != 0 {
		request.AddParam("limit", strconv.FormatInt(optionalRequest.EndTime, 10))
	}

	url := p.privateUrlBuilder.Build("GET", "/v2/account/ledger", request)
	getResp, getErr := internal.HttpGet(url)
	if getErr != nil {
		return nil, getErr
	}
	result := account.GetAccountLedgerResponse{}
	jsonErr := json.Unmarshal([]byte(getResp), &result)
	if jsonErr != nil {
		return nil, jsonErr
	}
	if result.Code == 200 && result.Data != nil {
		return result.Data, nil
	}

	return nil, errors.New(getResp)
}

// Transfer fund between spot account and future contract account
func (p *AccountClient) FuturesTransfer(request account.FuturesTransferRequest) (int64, error) {
	postBody, jsonErr := model.ToJson(request)
	if jsonErr != nil {
		return 0, jsonErr
	}

	url := p.privateUrlBuilder.Build("POST", "/v1/futures/transfer", nil)
	postResp, postErr := internal.HttpPost(url, postBody)
	if postErr != nil {
		return 0, postErr
	}

	result := account.FuturesTransferResponse{}
	jsonErr = json.Unmarshal([]byte(postResp), &result)
	if jsonErr != nil {
		return 0, jsonErr
	}
	if result.Status != "ok" {
		return 0, errors.New(postResp)

	}
	return result.Data, nil
}

// Returns the point balance of specified user's account
func (p *AccountClient) GetPointBalance(subUid string) (*account.GetPointBalanceResponse, error) {
	request := new(model.GetRequest).Init()
	request.AddParam("subUid", subUid)

	url := p.privateUrlBuilder.Build("GET", "/v2/point/account", request)
	getResp, getErr := internal.HttpGet(url)
	if getErr != nil {
		return nil, getErr
	}
	result := account.GetPointBalanceResponse{}
	jsonErr := json.Unmarshal([]byte(getResp), &result)
	if jsonErr != nil {
		return nil, jsonErr
	}
	if result.Code == 200 {
		return &result, nil
	}

	return nil, errors.New(getResp)
}

// Transfer points between spot account and future contract account
func (p *AccountClient) TransferPoint(request account.TransferPointRequest) (*account.TransferPointResponse, error) {
	postBody, jsonErr := model.ToJson(request)
	if jsonErr != nil {
		return nil, jsonErr
	}

	url := p.privateUrlBuilder.Build("POST", "/v2/point/transfer", nil)
	postResp, postErr := internal.HttpPost(url, postBody)
	if postErr != nil {
		return nil, postErr
	}

	result := account.TransferPointResponse{}
	jsonErr = json.Unmarshal([]byte(postResp), &result)
	if jsonErr != nil {
		return nil, jsonErr
	}
	if result.Code == 200 {
		return &result, nil
	}

	return nil, errors.New(postResp)
}
