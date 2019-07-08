// +build integration

package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/topfreegames/Will.IAM/models"
	helpers "github.com/topfreegames/Will.IAM/testing"
)

func beforeEachServices(t *testing.T) {
	t.Helper()
	storage := helpers.GetStorage(t)
	_, err := storage.PG.DB.Exec("TRUNCATE service_accounts CASCADE;")
	if err != nil {
		panic(err)
	}
}

func TestServicesCreateHandler(t *testing.T) {
	beforeEachServices(t)
	rootSA := helpers.CreateRootServiceAccount(t)
	saUC := helpers.GetServiceAccountsUseCase(t)
	sa := &models.ServiceAccount{
		Name:  "any",
		Email: "any@email.com",
	}
	if err := saUC.Create(sa); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}

	service := &models.Service{
		Name:                    "Some Service",
		PermissionName:          "SomeService",
		CreatorServiceAccountID: sa.ID,
		AMURL:                   "http://localhost:3333/am",
	}

	app := helpers.GetApp(t)
	bts, err := json.Marshal(service)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(bts))
	req.Header.Set("Authorization", fmt.Sprintf(
		"KeyPair %s:%s", rootSA.KeyID, rootSA.KeySecret,
	))
	rec := helpers.DoRequest(t, req, app.GetRouter())
	if rec.Code != http.StatusCreated {
		t.Errorf("Expected 201. Got %d", rec.Code)
		return
	}
	ssUC := helpers.GetServicesUseCase(t)
	ss, err := ssUC.List()
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	if len(ss) != 1 {
		t.Errorf("Expected to have 1 service. Got %d", len(ss))
		return
	}
	if ss[0].Name != "Some Service" {
		t.Errorf("Expected service name to be Some Service. Got %s", ss[0].Name)
		return
	}
	if ss[0].PermissionName != "SomeService" {
		t.Errorf(
			"Expected service permission name to be SomeService. Got %s",
			ss[0].PermissionName,
		)
		return
	}
}
