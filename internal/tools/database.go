package tools

import (
	log "github.com/sirupsen/logrus"
)

type LoginDetails struct {
	AuthToken string
	Username  string
}

type CoinDetails struct {
	Coins    int64
	Username string
}

type DatabaseInterface interface {
	GetUserLoginDetails(username string) *LoginDetails
	GetUserCoins(username string) *CoinDetails
	AddUserCoins(username string, amount int64) *CoinDetails
	WithdrawUserCoins(username string, amount int64) *CoinDetails
	TransferUserCoins(from string, to string, amount int64) (fromDetails *CoinDetails, toDetails *CoinDetails)
	SetupDatabase() error
}

func NewDatabase() (*DatabaseInterface, error) {
	var database DatabaseInterface = &mockDB{}
	var err error = database.SetupDatabase()
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return &database, nil
}
