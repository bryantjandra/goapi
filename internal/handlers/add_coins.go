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

func AddCoins(w http.ResponseWriter, r *http.Request) {
	//parse params
	var params = api.CoinAdditionParams{}
	var decoder *schema.Decoder = schema.NewDecoder()

	var err error = decoder.Decode(&params, r.URL.Query())

	if err != nil {
		log.Error("Failed to parse request parameters: ", err)
		api.RequestErrorHandler(w, err)
		return
	}

	//connect to DB
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

	//update the coin balance
	var updatedCoinBalance *tools.CoinDetails = (*database).AddUserCoins(params.Username, params.Amount)
	if updatedCoinBalance == nil {
		log.Error("Failed to add coins for user: ", params.Username)
		api.RequestErrorHandler(w, fmt.Errorf("user not found or invalid amount"))
		return
	}

	//return the response
	var response api.CoinAdditionResponse = api.CoinAdditionResponse{
		Code:    http.StatusOK,
		Message: "Your coin balance has been updated.",
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
