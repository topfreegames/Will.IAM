package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/ghostec/Will.IAM/errors"
	"github.com/ghostec/Will.IAM/models"
	"github.com/ghostec/Will.IAM/usecases"
	"github.com/gorilla/mux"
	"github.com/topfreegames/extensions/middleware"
)

func serviceAccountsGetHandler(
	sasUC usecases.ServiceAccounts,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l := middleware.GetLogger(r.Context())
		saID := mux.Vars(r)["id"]
		sawn, err := sasUC.WithContext(r.Context()).GetWithNested(saID)
		if err != nil {
			l.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		bts, err := json.Marshal(sawn)
		if err != nil {
			l.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		WriteBytes(w, 200, bts)
	}
}

func serviceAccountsCreateHandler(
	sasUC usecases.ServiceAccounts,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l := middleware.GetLogger(r.Context())
		sawn, err := processServiceAccountWithNestedFromReq(r, sasUC)
		if err != nil {
			statusCode := http.StatusInternalServerError
			if _, ok := err.(*errors.UserDoesntHaveAllPermissionsError); ok {
				statusCode = err.(*errors.UserDoesntHaveAllPermissionsError).
					StatusCode()
			}
			l.WithError(err).Error(
				"serviceAccountsCreateHandler processServiceAccountWithNestedFromReq",
			)
			w.WriteHeader(statusCode)
			return
		}
		v := sawn.Validate()
		if !v.Valid() {
			WriteBytes(w, http.StatusUnprocessableEntity, v.Errors())
			return
		}
		sawn.ID = mux.Vars(r)["id"]
		if err := sasUC.WithContext(r.Context()).CreateWithNested(sawn); err != nil {
			l.WithError(err).Error("sasUC.CreateWithNested failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	}
}

func serviceAccountsUpdateHandler(
	sasUC usecases.ServiceAccounts,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l := middleware.GetLogger(r.Context())
		sawn, err := processServiceAccountWithNestedFromReq(r, sasUC)
		if err != nil {
			statusCode := http.StatusInternalServerError
			if _, ok := err.(*errors.UserDoesntHaveAllPermissionsError); ok {
				statusCode = err.(*errors.UserDoesntHaveAllPermissionsError).
					StatusCode()
			}
			l.WithError(err).Error(
				"serviceAccountsUpdateHandler processServiceAccountWithNestedFromReq",
			)
			w.WriteHeader(statusCode)
			return
		}
		v := sawn.Validate()
		if !v.Valid() {
			WriteBytes(w, http.StatusUnprocessableEntity, v.Errors())
			return
		}
		sawn.ID = mux.Vars(r)["id"]
		if err := sasUC.WithContext(r.Context()).UpdateWithNested(sawn); err != nil {
			l.WithError(err).Error("sasUC.UpdateWithNested failed")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}

func processServiceAccountWithNestedFromReq(
	r *http.Request, sasUC usecases.ServiceAccounts,
) (*usecases.ServiceAccountWithNested, error) {
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return nil, err
	}
	sawn := &usecases.ServiceAccountWithNested{}
	err = json.Unmarshal(body, sawn)
	if err != nil {
		return nil, err
	}
	saID, _ := getServiceAccountID(r.Context())
	uc := sasUC.WithContext(r.Context())
	hasAllOwnerRolesPermissions, err :=
		uc.HasAllOwnerRolesPermissions(saID, sawn.RolesIDs)
	if err != nil {
		return nil, err
	}
	if !hasAllOwnerRolesPermissions {
		return nil, errors.NewUserDoesntHaveAllPermissionsError()
	}
	sawn.Permissions, err = models.BuildPermissions(sawn.PermissionsStrings)
	if err != nil {
		return nil, err
	}
	for i := range sawn.PermissionsStrings {
		if alias, ok := sawn.PermissionsAliases[sawn.PermissionsStrings[i]]; ok {
			sawn.Permissions[i].Alias = alias
		}
	}
	has, err := uc.HasAllOwnerPermissions(saID, sawn.Permissions)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, errors.NewUserDoesntHaveAllPermissionsError()
	}
	return sawn, nil
}

func serviceAccountsListHandler(
	sasUC usecases.ServiceAccounts,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l := middleware.GetLogger(r.Context())
		listOptions, err := buildListOptions(r)
		if err != nil {
			Write(
				w, http.StatusUnprocessableEntity,
				fmt.Sprintf(`{ "error": "%s"  }`, err.Error()),
			)
			return
		}
		saSl, count, err := sasUC.WithContext(r.Context()).List(listOptions)
		if err != nil {
			l.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		results, err := keepJSONFields(
			saSl, "id", "authenticationType", "name", "email", "picture",
		)
		if err != nil {
			l.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		ret := map[string]interface{}{
			"count":   count,
			"results": results,
		}
		WriteJSON(w, 200, ret)
	}
}

func serviceAccountsSearchHandler(
	sasUC usecases.ServiceAccounts,
) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		l := middleware.GetLogger(r.Context())
		term := r.URL.Query().Get("term")
		listOptions, err := buildListOptions(r)
		if err != nil {
			Write(
				w, http.StatusUnprocessableEntity,
				fmt.Sprintf(`{ "error": "%s"  }`, err.Error()),
			)
			return
		}
		saSl, count, err := sasUC.WithContext(r.Context()).Search(term, listOptions)
		if err != nil {
			l.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		results, err := keepJSONFields(
			saSl, "id", "authenticationType", "name", "email", "picture", "baseRoleId",
		)
		if err != nil {
			l.Error(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		ret := map[string]interface{}{
			"count":   count,
			"results": results,
		}
		WriteJSON(w, 200, ret)
	}
}
