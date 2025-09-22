package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/bryantjandra/goapi/api"
	"github.com/bryantjandra/goapi/internal/tools"
	"github.com/gorilla/schema"
	log "github.com/sirupsen/logrus"
)

func WithdrawCoins(w http.ResponseWriter, r *http.Request) {
	//parse params
	var params = api.CoinWithdrawParams{}
	var decoder *schema.Decoder = schema.NewDecoder()

	var err error = decoder.Decode(&params, r.URL.Query())

	if err != nil {
		log.Error("Failed to parse request parameters: ", err)
		api.RequestErrorHandler(w, err)
		return
	}

	var database *tools.DatabaseInterface
	database, err = tools.NewDatabase()
	if err != nil {
		log.Error("Failed to connect to database: ", err)
		api.InternalErrorHandler(w)
		return
	}

	// Validate amount is positive
	if params.Amount <= 0 {
		log.Error("Invalid amount: must be positive, got: ", params.Amount)
		api.RequestErrorHandler(w, fmt.Errorf("amount must be positive"))
		return
	}

	// Get original balance before withdrawal
	var originalBalance *tools.CoinDetails = (*database).GetUserCoins(params.Username)
	if originalBalance == nil {
		log.Error("User not found: ", params.Username)
		api.RequestErrorHandler(w, fmt.Errorf("user not found"))
		return
	}

	var updatedCoinBalance *tools.CoinDetails = (*database).WithdrawUserCoins(params.Username, params.Amount)
	if updatedCoinBalance == nil {
		log.Error("Withdrawal failed for user: ", params.Username, " amount: ", params.Amount)
		api.RequestErrorHandler(w, fmt.Errorf("insufficient funds or invalid amount"))
		return
	}

	var response api.CoinWithdrawResponse = api.CoinWithdrawResponse{
		Code:    200,
		Message: fmt.Sprintf("You have successfully withdrawn %d. Your original coin balance was %d, now it is %d", params.Amount, originalBalance.Coins, updatedCoinBalance.Coins),
		Amount:  params.Amount,
		Balance: updatedCoinBalance.Coins,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		log.Error("Failed to encode response: ", err)
		api.InternalErrorHandler(w)
		return
	}

}
