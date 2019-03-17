package controlflow

import (
	"fmt"
	"log"

	"github.com/dk1027/go-questrade-api/api"
)

type Checker struct {
	Session *api.Session
}

func CHECK(e error, errMsg string) {
	if e != nil {
		log.Fatalln(errMsg)
	}
}

func NewChecker(refreshToken string) *Checker {
	session, err := api.Redeem(refreshToken)
	CHECK(err, "Error redeeming refresh token")
	return &Checker{session}
}

type Portfolio []LineItem

type LineItem struct {
	Account string  `json:"Account"`
	Symbol  string  `json:"Symbol"`
	Amount  float64 `json:"Amount"`
}

func (l LineItem) String() string {
	return fmt.Sprintf("%s, %s, %v", l.Account, l.Symbol, l.Amount)
}

func (c *Checker) Get() Portfolio {
	var portfolio Portfolio
	accounts, err := api.Accounts(c.Session)
	CHECK(err, "Error getting accounts")
	for _, account := range accounts.Accounts {
		balances, _ := api.Balances(c.Session, account.Number)

		for _, balance := range balances.PerCurrencyBalances {
			portfolio = append(portfolio, LineItem{account.Number, "CASH", balance.Cash})
		}

		positions, _ := api.Positions(c.Session, account.Number)
		for _, position := range positions.Positions {
			portfolio = append(portfolio, LineItem{account.Number, position.Symbol, position.CurrentMarketValue})
		}
	}
	for _, line := range portfolio {
		log.Println(line)
	}
	return portfolio
}
