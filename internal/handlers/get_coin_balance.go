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

func GetCoinBalance(w http.ResponseWriter, r *http.Request) {
	var params = api.CoinBalanceParams{}
	var decoder *schema.Decoder = schema.NewDecoder()
	var err error

	err = decoder.Decode(&params, r.URL.Query())

	if err != nil {
		log.Error("Failed to parse request parameters: ", err)
		api.RequestErrorHandler(w, err)
		return
	}

	database, err := tools.NewDatabase()
	if err != nil {
		log.Error("Failed to connect to database: ", err)
		api.InternalErrorHandler(w)
		return
	}

	tokenDetails := (*database).GetUserCoins(params.Username)
	if tokenDetails == nil {
		log.Error("User not found: ", params.Username)
		api.RequestErrorHandler(w, fmt.Errorf("user not found"))
		return
	}

	var response = api.CoinBalanceResponse{
		Balance: (*tokenDetails).Coins,
		Code:    http.StatusOK,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Error("Failed to encode response: ", err)
		api.InternalErrorHandler(w)
		return
	}

}
