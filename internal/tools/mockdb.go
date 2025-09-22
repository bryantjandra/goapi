package tools

import (
	"time"
)

type mockDB struct{}

// Mock login details database
var mockLoginDetails = map[string]LoginDetails{
	"aaron": {
		AuthToken: "1",
		Username:  "aaron",
	},
	"bryan": {
		AuthToken: "2",
		Username:  "jason",
	},
}

// Mock coin balance database
var mockCoinDetails = map[string]CoinDetails{
	"aaron": {
		Coins:    1000,
		Username: "aaron",
	},
	"bryan": {
		Coins:    1000,
		Username: "bryan",
	},
}

func (d *mockDB) GetUserLoginDetails(username string) *LoginDetails {
	// Simulate DB call
	time.Sleep(time.Second * 1)

	var clientData = LoginDetails{}
	clientData, ok := mockLoginDetails[username]
	if !ok {
		return nil
	}

	return &clientData
}

func (d *mockDB) GetUserCoins(username string) *CoinDetails {
	// Simulate DB call
	time.Sleep(time.Second * 1)

	var clientData = CoinDetails{}
	clientData, ok := mockCoinDetails[username]
	if !ok {
		return nil
	}

	return &clientData
}

func (d *mockDB) SetupDatabase() error {
	return nil
}

func (d *mockDB) AddUserCoins(username string, amount int64) *CoinDetails {
	// Simulate DB call
	time.Sleep(time.Second * 1)

	var clientData = CoinDetails{}
	clientData, ok := mockCoinDetails[username]
	if !ok {
		return nil
	}

	// update the coins
	clientData.Coins = clientData.Coins + amount

	// save changes back to the mock datbase
	mockCoinDetails[username] = clientData

	return &clientData
}

func (d *mockDB) WithdrawUserCoins(username string, amount int64) *CoinDetails {
	// Simulate DB call
	time.Sleep(time.Second * 1)

	var clientData = CoinDetails{}
	clientData, ok := mockCoinDetails[username]
	if !ok {
		return nil
	}

	if amount > clientData.Coins {
		return nil
	}

	// decrement the coin balance
	clientData.Coins = clientData.Coins - amount

	// save changes back to mock db
	mockCoinDetails[username] = clientData

	return &clientData
}

func (d *mockDB) TransferUserCoins(from string, to string, amount int64) (fromDetails *CoinDetails, toDetails *CoinDetails) {
	// Simulate DB call
	time.Sleep(time.Second * 1)

	var fromData = CoinDetails{}
	fromData, ok := mockCoinDetails[from]
	if !ok {
		return nil, nil
	}

	var toData = CoinDetails{}
	toData, okTwo := mockCoinDetails[to]
	if !okTwo {
		return nil, nil
	}

	// check sufficient balance
	if fromData.Coins < amount {
		return nil, nil
	}

	//update sender's and receiver's balance
	fromData.Coins = fromData.Coins - amount
	mockCoinDetails[from] = fromData

	toData.Coins = toData.Coins + amount
	mockCoinDetails[to] = toData

	return &fromData, &toData

}
