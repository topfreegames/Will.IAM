package api

import (
	"fmt"
	"net/http"

	"github.com/topfreegames/Will.IAM/models"
	"github.com/topfreegames/Will.IAM/usecases"
	"github.com/topfreegames/extensions/middleware"
)

func permissionsRequestsCreateHandler(
	prsUC usecases.PermissionsRequests,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l := middleware.GetLogger(r.Context())
		pr := &models.PermissionRequest{}
		if err := readBodyTo(r, pr); err != nil {
			l.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		saID, _ := getServiceAccountID(r.Context())
		pr.ServiceAccountID = saID
		if err := prsUC.WithContext(r.Context()).Create(pr); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if pr.ID == "" {
			w.WriteHeader(http.StatusAccepted)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func permissionsRequestsListOpenHandler(
	prsUC usecases.PermissionsRequests,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l := middleware.GetLogger(r.Context())
		saID, _ := getServiceAccountID(r.Context())
		listOptions, err := buildListOptions(r)
		if err != nil {
			Write(
				w, http.StatusUnprocessableEntity,
				fmt.Sprintf(`{ "error": "%s"  }`, err.Error()),
			)
			return
		}
		prs, count, err := prsUC.WithContext(r.Context()).ListOpenRequestsVisibleTo(listOptions, saID)
		if err != nil {
			l.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		WriteJSON(w, 200, map[string]interface{}{
			"count":   count,
			"results": prs,
		})
	}
}
