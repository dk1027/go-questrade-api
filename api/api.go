package api

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/levigross/grequests"
)

type Session struct {
	AccessToken  string `json:"access_token"`
	ApiServer    string `json:"api_server"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
}

func makeRedeemResponse(data string) (*Session, error) {
	resp := &Session{}
	err := json.Unmarshal([]byte(data), resp)
	if err != nil {
		log.Fatalln("Error making Session: ", data)
		return resp, err
	}
	strings.Replace(resp.ApiServer, "\\/", "/", -1)
	return resp, nil
}

func Redeem(refreshToken string) (*Session, error) {
	params := &grequests.RequestOptions{
		Params: map[string]string{
			"grant_type":    "refresh_token",
			"refresh_token": refreshToken}}

	resp, err := grequests.Post("https://login.questrade.com/oauth2/token", params)
	if err != nil {
		log.Fatalln("Unable to make request: ", err)
		log.Fatalln("Received status code", resp.StatusCode)
		log.Fatalln(resp.String())
		return nil, err
	}

	if resp.StatusCode == 200 {
		//log.Println(resp.String())
		result, err := makeRedeemResponse(resp.String())
		return result, err
	}
	log.Println(resp.String())
	log.Fatalln("Unable to Redeem refresh token. Status Code is not 200", resp.StatusCode)
	return nil, err
}

type ApiError struct{}

func (e *ApiError) Error() string {
	return "Request failed"
}

type Account struct {
	Number string `json:"number"`
}

type AccountsResponse struct {
	Accounts []Account `json:"accounts"`
}

type Position struct {
	Symbol             string  `json:"symbol"`
	CurrentMarketValue float64 `json:"currentMarketValue"`
	CurrentPrice       float64 `json:"currentPrint"`
}

type PositionsResponse struct {
	Positions []Position `json:"positions"`
}

type Balance struct {
	Currency string  `json:"currency"`
	Cash     float64 `json:"cash"`
}

type BalancesResponse struct {
	PerCurrencyBalances []Balance `json:"perCurrencyBalances"`
}

func get(session *Session, endpoint string) (*grequests.Response, error) {
	params := &grequests.RequestOptions{
		Headers: map[string]string{"Authorization": "Bearer " + session.AccessToken},
	}
	url := session.ApiServer + endpoint
	log.Println(url)
	resp, err := grequests.Get(session.ApiServer+endpoint, params)
	return resp, err
}

func Accounts(session *Session) (*AccountsResponse, error) {
	result := &AccountsResponse{}
	endpoint := "v1/accounts"
	retried := false
Retry:
	resp, err := get(session, endpoint)
	if err != nil {
		log.Fatalf("%v failed: %v\n", endpoint, err)
		return nil, err
	}
	if resp.StatusCode == 401 {
		newSession, err := Redeem(session.RefreshToken)
		*session = *newSession
		if err != nil {
			log.Println(err)
			log.Fatalf("Unable to redeem refresh token %v", err)
		}
		if !retried {
			retried = true
			goto Retry
		}

	}
	if resp.StatusCode != 200 {
		log.Println(resp.String())
		log.Fatalf("%v status code is %d", endpoint, resp.StatusCode)
		return nil, &ApiError{}
	}
	err = json.Unmarshal([]byte(resp.String()), result)
	if err != nil {
		log.Fatalln("Failed to parse result.", err)
		return nil, err
	}
	return result, nil
}

func CheckStatus(statusCode int) {
	if statusCode != 200 {
		log.Fatalf("status code is %d", statusCode)
	}
}

func CheckError(err error, msg string) {
	if err != nil {
		log.Fatalf("%v %v", msg, err)
	}
}

func CheckHttpResponse(err error, msg string) {
	if err != nil {
		log.Fatalf("%v %v\n", msg, err)
	}
}

func Positions(session *Session, id string) (*PositionsResponse, error) {
	result := &PositionsResponse{}
	endpoint := fmt.Sprintf("v1/accounts/%v/positions", id)
	resp, err := get(session, endpoint)

	CheckHttpResponse(err, endpoint)

	CheckStatus(resp.StatusCode)

	err = json.Unmarshal([]byte(resp.String()), result)
	CheckError(err, "Failed to parse result.")

	return result, nil
}

func Balances(session *Session, id string) (*BalancesResponse, error) {
	result := &BalancesResponse{}
	endpoint := fmt.Sprintf("v1/accounts/%v/balances", id)
	resp, err := get(session, endpoint)

	CheckHttpResponse(err, endpoint)

	CheckStatus(resp.StatusCode)

	err = json.Unmarshal([]byte(resp.String()), result)
	CheckError(err, "Failed to parse result.")

	return result, nil
}
