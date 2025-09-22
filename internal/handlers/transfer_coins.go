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

func TransferCoins(w http.ResponseWriter, r *http.Request) {
	//parse params
	var params = api.CoinTransferParams{}
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

	// Validate username matches from parameter for security
	if params.Username != params.From {
		log.Error("Security violation: username doesn't match from parameter")
		api.RequestErrorHandler(w, fmt.Errorf("cannot transfer from another user's account"))
		return
	}

	fromDetails, toDetails := (*database).TransferUserCoins(params.From, params.To, params.Amount)
	if fromDetails == nil || toDetails == nil {
		log.Error("Transfer failed for users: ", params.From, " -> ", params.To, " amount: ", params.Amount)
		api.RequestErrorHandler(w, fmt.Errorf("transfer failed: user not found, insufficient funds, or invalid parameters"))
		return
	}

	var response api.CoinTransferResponse = api.CoinTransferResponse{
		Code:        200,
		Message:     fmt.Sprintf("You have successfully transferred %d to %s. Your current balance is %d", params.Amount, params.To, fromDetails.Coins),
		FromBalance: fromDetails.Coins,
		ToBalance:   toDetails.Coins,
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)

	if err != nil {
		log.Error("Failed to encode response: ", err)
		api.InternalErrorHandler(w)
		return
	}

}
