package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/topfreegames/Will.IAM/errors"
	"github.com/topfreegames/Will.IAM/models"
	"github.com/topfreegames/Will.IAM/usecases"
	"github.com/topfreegames/extensions/middleware"
)

func permissionsDeleteHandler(
	sasUC usecases.ServiceAccounts, psUC usecases.Permissions,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l := middleware.GetLogger(r.Context())
		pID := mux.Vars(r)["id"]
		p, err := psUC.WithContext(r.Context()).Get(pID)
		if err != nil {
			if _, ok := err.(*errors.EntityNotFoundError); ok {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		p.OwnershipLevel = models.OwnershipLevels.Owner
		saID, _ := getServiceAccountID(r.Context())
		has, err := sasUC.WithContext(r.Context()).HasPermissionString(saID, p.String())
		if err != nil {
			l.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !has {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		err = psUC.WithContext(r.Context()).Delete(pID)
		if err != nil {
			l.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func permissionsAttributeHandler(
	sasUC usecases.ServiceAccounts, psUC usecases.Permissions,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l := middleware.GetLogger(r.Context())
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			l.WithError(err).Error("ioutil.ReadAll failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		pa := &usecases.PermissionsAttribute{}
		err = json.Unmarshal(body, pa)
		if err != nil {
			l.WithError(err).Error("json.Unmarshal failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		pa.Permissions, err = models.BuildPermissions(pa.PermissionsStrings)
		if err != nil {
			l.WithError(err).Error("BuildPermissions failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		for i := range pa.PermissionsStrings {
			if alias, ok := pa.PermissionsAliases[pa.PermissionsStrings[i]]; ok {
				pa.Permissions[i].Alias = alias
			}
		}
		saID, _ := getServiceAccountID(r.Context())
		has, err := sasUC.WithContext(r.Context()).
			HasAllOwnerPermissions(saID, pa.Permissions)
		if err != nil {
			l.WithError(err).Error("HasAllOwnerPermissions failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !has {
			l.Infof("saID %s doesn't own all permissions")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		err = psUC.WithContext(r.Context()).Attribute(pa)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func permissionsAttributeToEmailsHandler(
	sasUC usecases.ServiceAccounts, psUC usecases.Permissions,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l := middleware.GetLogger(r.Context())
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			l.WithError(err).Error("ioutil.ReadAll failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		pa := &usecases.PermissionsAttributeToEmails{}
		err = json.Unmarshal(body, pa)
		if err != nil {
			l.WithError(err).Error("json.Unmarshal failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		pa.Permissions, err = models.BuildPermissions(pa.PermissionsStrings)
		if err != nil {
			l.WithError(err).Error("BuildPermissions failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		for i := range pa.PermissionsStrings {
			if alias, ok := pa.PermissionsAliases[pa.PermissionsStrings[i]]; ok {
				pa.Permissions[i].Alias = alias
			}
		}
		saID, _ := getServiceAccountID(r.Context())
		has, err := sasUC.WithContext(r.Context()).
			HasAllOwnerPermissions(saID, pa.Permissions)
		if err != nil {
			l.WithError(err).Error("HasAllOwnerPermissions failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !has {
			l.Infof("saID %s doesn't own all permissions")
			w.WriteHeader(http.StatusForbidden)
			return
		}
		err = psUC.WithContext(r.Context()).AttributeToEmails(pa)
		if err != nil {
			l.WithError(err).Error("AttributeToEmails failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func permissionsHasHandler(
	sasUC usecases.ServiceAccounts,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l := middleware.GetLogger(r.Context())
		qs := r.URL.Query()
		permissionSl := qs["permission"]
		if len(permissionSl) == 0 {
			Write(w, http.StatusUnprocessableEntity,
				`{"error": "querystrings.permission is required"}`)
			return
		}
		saID, _ := getServiceAccountID(r.Context())
		has, err :=
			sasUC.WithContext(r.Context()).HasPermissionString(saID, permissionSl[0])
		if err != nil {
			l.Error(err)
			Write(w, http.StatusUnprocessableEntity,
				`{"error": "Incomplete permission. Expected format: Service::OwnershipLevel::Action::{ResourceHierarchy}"}`)
			return
		}
		if !has {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func permissionsHasManyHandler(
	sasUC usecases.ServiceAccounts,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l := middleware.GetLogger(r.Context())
		body, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		if err != nil {
			l.WithError(err).Error("permissionsHasMany ioutil.ReadAll failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		permissions := make([]string, 0)
		err = json.Unmarshal(body, &permissions)
		if err != nil {
			l.WithError(err).Error("permissionsHasMany json.Unmarshal failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		saID, _ := getServiceAccountID(r.Context())
		resultStatus, err :=
			sasUC.WithContext(r.Context()).HasPermissionsStrings(saID, permissions)
		if err != nil {
			l.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		bts, err := json.Marshal(resultStatus)
		if err != nil {
			l.WithError(err).Error("permissionsHasMany json.Marshal error")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		WriteBytes(w, http.StatusOK, bts)
	}
}
