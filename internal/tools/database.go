package tools

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"
)

type LoginDetails struct {
	AuthToken string
	Username  string
}

type CoinDetails struct {
	Coins    int64
	Username string
	Version  int64 // Optimistic locking
}

// Transaction audit trail
type TransactionLog struct {
	ID        string
	Type      string
	From      string
	To        string
	Amount    int64
	Timestamp time.Time
	Status    string
}

type DatabaseInterface interface {
	GetUserLoginDetails(username string) *LoginDetails
	GetUserCoins(username string) *CoinDetails
	AddUserCoins(username string, amount int64) *CoinDetails
	WithdrawUserCoins(username string, amount int64) *CoinDetails
	TransferUserCoins(from string, to string, amount int64) (fromDetails *CoinDetails, toDetails *CoinDetails)
	SetupDatabase() error
	TransferUserCoinsWithContext(ctx context.Context, from string, to string, amount int64) (fromDetails *CoinDetails, toDetails *CoinDetails, err error)
	GetTransactionHistory(username string) []TransactionLog
	GetSystemHealth() map[string]interface{}
}

func NewDatabase() (*DatabaseInterface, error) {
	log.Debug("Creating new database connection")

	var database DatabaseInterface = &mockDB{}
	var err error = database.SetupDatabase()
	if err != nil {
		log.Error("Failed to setup database: ", err)
		return nil, err
	}

	log.Debug("Database connection established successfully")
	return &database, nil
}
