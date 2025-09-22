package middleware

import (
	"errors"
	"net/http"

	"github.com/bryantjandra/goapi/api"
	"github.com/bryantjandra/goapi/internal/tools"
	log "github.com/sirupsen/logrus"
)

var UnAuthorizedError = errors.New("Invalid username or token")

func Authorization(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var username string = r.URL.Query().Get("username")
		var token = r.Header.Get("Authorization")

		if username == "" || token == "" {
			log.Error("Authorization failed: missing username or token")
			api.RequestErrorHandler(w, UnAuthorizedError)
			return
		}

		database, err := tools.NewDatabase()
		if err != nil {
			log.Error("Failed to connect to database during authorization: ", err)
			api.InternalErrorHandler(w)
			return
		}

		loginDetails := (*database).GetUserLoginDetails(username)

		if loginDetails == nil || (token != (*loginDetails).AuthToken) {
			log.Error("Authorization failed for user: ", username, " - invalid credentials")
			api.RequestErrorHandler(w, UnAuthorizedError)
			return
		}

		next.ServeHTTP(w, r)
	})
}
