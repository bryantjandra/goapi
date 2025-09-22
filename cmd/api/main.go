package main

import (
	"fmt"
	"net/http"

	"github.com/bryantjandra/goapi/internal/handlers"
	"github.com/go-chi/chi"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetReportCaller(true)

	log.Info("Initializing GO API Service...")

	var r *chi.Mux = chi.NewRouter()
	handlers.Handler(r)

	fmt.Println("Starting GO API Service...")
	log.Info("Server starting on localhost:3000")

	err := http.ListenAndServe("localhost:3000", r)
	if err != nil {
		log.Fatal("Failed to start server: ", err)
	}
}
