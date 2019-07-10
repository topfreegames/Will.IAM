package api

import (
	"net/http"

	"github.com/topfreegames/Will.IAM/usecases"
)

func healthcheckHandler(
	uc usecases.Healthcheck,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := uc.Do(); err != nil {
			Write(w, http.StatusInternalServerError, `{"healthy": false}`)
		} else {
			Write(w, http.StatusOK, `{"healthy": true}`)
		}
	}
}
