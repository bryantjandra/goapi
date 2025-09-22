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

	fromDetails, toDetails := (*database).TransferUserCoins(params.From, params.To, params.Amount)
	if fromDetails == nil || toDetails == nil {
		log.Error("Transfer failed.")
		api.InternalErrorHandler(w)
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
		log.Error(err)
		api.InternalErrorHandler(w)
		return
	}

}
