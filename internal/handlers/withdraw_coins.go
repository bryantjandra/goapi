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
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	var database *tools.DatabaseInterface
	database, err = tools.NewDatabase()
	if err != nil {
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

	// Get original balance before withdrawal
	var originalBalance *tools.CoinDetails = (*database).GetUserCoins(params.Username)
	if originalBalance == nil {
		log.Error("User not found")
		api.InternalErrorHandler(w)
		return
	}

	var updatedCoinBalance *tools.CoinDetails = (*database).WithdrawUserCoins(params.Username, params.Amount)
	if updatedCoinBalance == nil {
		log.Error(err)
		api.InternalErrorHandler(w)
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
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

}
