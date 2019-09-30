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
	rootSA := helpers.CreateRootServiceAccountWithKeyPair(t, "rootSAKeyPair", "rootSAKeyPair@test.com")
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

func TestServicesGetHandler(t *testing.T) {
	beforeEachServices(t)

	rootSA := helpers.CreateRootServiceAccountWithKeyPair(t, "rootSAKeyPair", "rootSAKeyPair@test.com")
	service := &models.Service{
		Name:                    "Some Service",
		PermissionName:          "SomeService",
		CreatorServiceAccountID: rootSA.ID,
		AMURL:                   "http://localhost:3333/am",
	}
	servicesUC := helpers.GetServicesUseCase(t)
	err := servicesUC.Create(service)
	if err != nil {
		t.Fatalf("Error persisting Service = %v", service)
	}

	testCases := []struct {
		name       string
		id         string
		wantStatus int
	}{
		{
			name:       "BlankID",
			id:         "",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "InexistentID",
			id:         "e6fee046-6045-45d2-b6f1-a21b82977782",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "InvalidUUID",
			id:         "o2206115-4f58-4d44-b200-5b227098070a",
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "FoundID",
			id:         service.ID,
			wantStatus: http.StatusOK,
		},
	}

	app := helpers.GetApp(t)
	bts, err := json.Marshal(service)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("/services/%s", testCase.id), bytes.NewBuffer(bts))
			req.Header.Set("Authorization", fmt.Sprintf("KeyPair %s:%s", rootSA.KeyID, rootSA.KeySecret))

			rec := helpers.DoRequest(t, req, app.GetRouter())

			if rec.Code != testCase.wantStatus {
				t.Errorf("Want HTTP Status %v, got %v", testCase.wantStatus, rec.Code)
			}
		})
	}
}
