package api

import (
	"net/http"

	"github.com/gorilla/mux"
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
		if err := unmarshalBodyTo(r, pr); err != nil {
			l.WithError(err).Error("failed to read body")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		saID, _ := getServiceAccountID(r.Context())
		pr.ServiceAccountID = saID
		if err := prsUC.WithContext(r.Context()).Create(pr); err != nil {
			l.WithError(err).Error("failed to create permission request")
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

func permissionsRequestsDenyHandler(
	prsUC usecases.PermissionsRequests,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l := middleware.GetLogger(r.Context())
		saID, _ := getServiceAccountID(r.Context())
		prID := mux.Vars(r)["id"]
		if err := prsUC.WithContext(r.Context()).Deny(saID, prID); err != nil {
			l.WithError(err).Error("failed to deny permission request")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}
}

func permissionsRequestsGrantHandler(
	prsUC usecases.PermissionsRequests,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l := middleware.GetLogger(r.Context())
		saID, _ := getServiceAccountID(r.Context())
		prID := mux.Vars(r)["id"]
		if err := prsUC.WithContext(r.Context()).Grant(saID, prID); err != nil {
			l.WithError(err).Error("failed to grant permission request")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusAccepted)
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
			WriteJSON(w, http.StatusUnprocessableEntity, ErrorResponse{Error: err.Error()})
			return
		}
		prs, count, err := prsUC.WithContext(r.Context()).ListOpenRequestsVisibleTo(listOptions, saID)
		if err != nil {
			l.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		WriteJSON(w, 200, ListResponse{Count: count, Results: prs})
	}
}
