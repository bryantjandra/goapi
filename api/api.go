package api

import (
	"encoding/json"
	"net/http"
)

// Coin Balance Params
type CoinBalanceParams struct {
	Username string
}

// Coin Balance Response
type CoinBalanceResponse struct {
	// Success Code, usually 200
	Code int

	// Account Balance
	Balance int64
}

type CoinAdditionParams struct {
	Username string
	Amount   int64
}

type CoinAdditionResponse struct {
	Code    int
	Message string
	Balance int64
}

type CoinWithdrawParams struct {
	Username string
	Amount   int64
}

type CoinWithdrawResponse struct {
	Code    int
	Message string
	Amount  int64
	Balance int64
}

// Error Response
type Error struct {
	// Error Code
	Code int

	// Error message
	Message string
}

func writeError(w http.ResponseWriter, message string, code int) {
	resp := Error{
		Code:    code,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	json.NewEncoder(w).Encode(resp)
}

var (
	RequestErrorHandler = func(w http.ResponseWriter, err error) {
		writeError(w, err.Error(), http.StatusBadRequest)
	}
	InternalErrorHandler = func(w http.ResponseWriter) {
		writeError(w, "An unexpected error occured.", http.StatusInternalServerError)
	}
)
