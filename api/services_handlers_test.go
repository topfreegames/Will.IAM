// +build integration

package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/kylelemons/godebug/pretty"
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
		t.Fatalf("Unexpected error: %v", err)
	}

	service := &models.Service{
		Name:                    "Some Service",
		PermissionName:          "SomeService",
		CreatorServiceAccountID: sa.ID,
		AMURL:                   "http://localhost:3333/am",
	}

	invalidService := &models.Service{
		Name:                    "",
		PermissionName:          "SomeService",
		CreatorServiceAccountID: sa.ID,
		AMURL:                   "http://localhost:3333/am",
	}

	validServiceJSON, err := json.Marshal(service)
	if err != nil {
		t.Fatalf("Marshal() returned err = %v", err)
	}

	invalidServiceJSON, err := json.Marshal(invalidService)
	if err != nil {
		t.Fatalf("Marshal() returned err = %v", err)
	}

	testCases := []struct {
		name       string
		json       []byte
		wantStatus int
	}{
		{
			name:       "MalformedJSON",
			json:       []byte("{"),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "InvalidJSON",
			json:       invalidServiceJSON,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "ValidJSON",
			json:       validServiceJSON,
			wantStatus: http.StatusCreated,
		},
	}

	app := helpers.GetApp(t)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(testCase.json))
			req.Header.Set("Authorization", fmt.Sprintf("KeyPair %v:%v", rootSA.KeyID, rootSA.KeySecret))

			resp := helpers.DoRequest(t, req, app.GetRouter())
			if resp.Code != testCase.wantStatus {
				t.Fatalf("Code = %v, want %v", resp.Code, testCase.wantStatus)
			}
		})
	}
}

func TestServicesCreateHandler_servicePersisted(t *testing.T) {
	beforeEachServices(t)

	rootSA := helpers.CreateRootServiceAccountWithKeyPair(t, "rootSAKeyPair", "rootSAKeyPair@test.com")
	saUC := helpers.GetServiceAccountsUseCase(t)
	sa := &models.ServiceAccount{
		Name:  "any",
		Email: "any@email.com",
	}
	if err := saUC.Create(sa); err != nil {
		t.Fatalf("Create() returned error = %v", err)
	}

	service := &models.Service{
		Name:                    "Some Service",
		PermissionName:          "SomeService",
		CreatorServiceAccountID: sa.ID,
		AMURL:                   "http://localhost:3333/am",
	}

	reqJSON, err := json.Marshal(service)
	if err != nil {
		t.Fatalf("Marshal() returned error = %v", err)
	}

	req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(reqJSON))
	req.Header.Set("Authorization", fmt.Sprintf("KeyPair %s:%s", rootSA.KeyID, rootSA.KeySecret))

	app := helpers.GetApp(t)
	resp := helpers.DoRequest(t, req, app.GetRouter())
	if resp.Code != http.StatusCreated {
		t.Fatalf("Code = %v, want %v", resp.Code, http.StatusCreated)
	}

	servicesUC := helpers.GetServicesUseCase(t)
	services, err := servicesUC.List()
	if err != nil {
		t.Fatalf("List() returned error = %v", err)
	}

	got := services[0]
	want := service
	// Fields set during Service creation
	want.ID = got.ID
	want.ServiceAccountID = got.ServiceAccountID
	want.CreatedAt = got.CreatedAt
	want.UpdatedAt = got.UpdatedAt
	want.CreatorServiceAccountID = got.CreatorServiceAccountID

	if diff := pretty.Compare(got, want); diff != "" {
		t.Errorf("%s: Services diff: (-got +want)\n%s", got, diff)
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
		t.Fatalf("Create() returned error = %v", service)
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
			id:         "x",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "FoundID",
			id:         service.ID,
			wantStatus: http.StatusOK,
		},
	}

	app := helpers.GetApp(t)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", fmt.Sprintf("/services/%s", testCase.id), bytes.NewBuffer([]byte("")))
			req.Header.Set("Authorization", fmt.Sprintf("KeyPair %s:%s", rootSA.KeyID, rootSA.KeySecret))

			resp := helpers.DoRequest(t, req, app.GetRouter())
			if resp.Code != testCase.wantStatus {
				t.Errorf("HTTP Status = %v, want %v", resp.Code, testCase.wantStatus)
			}
		})
	}
}

func TestServicesUpdateHandler(t *testing.T) {
	beforeEachServices(t)

	rootSA := helpers.CreateRootServiceAccountWithKeyPair(t, "rootSAKeyPair", "rootSAKeyPair@test.com")
	service := &models.Service{
		Name:                    "Some Service",
		PermissionName:          "SomeService",
		CreatorServiceAccountID: rootSA.ID,
		AMURL:                   "http://localhost:3333/am",
	}
	servicesUC := helpers.GetServicesUseCase(t)
	if err := servicesUC.Create(service); err != nil {
		t.Fatalf("Create() returned error = %v", err)
	}

	service.Name = "Another name"
	validJSON, err := json.Marshal(service)
	if err != nil {
		t.Fatalf("Marshal() returned error = %v", err)
	}

	service.Name = ""
	invalidJSON, err := json.Marshal(service)
	if err != nil {
		t.Fatalf("Marshal() returned error = %v", err)
	}

	app := helpers.GetApp(t)

	testCases := []struct {
		name       string
		id         string
		json       []byte
		wantStatus int
	}{
		{
			name:       "InexistentID",
			id:         "e6fee046-6045-45d2-b6f1-a21b82977782",
			json:       validJSON,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "InvalidUUID",
			id:         "x",
			json:       validJSON,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "MalformedJSON",
			id:         service.ID,
			json:       []byte("{"),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "InvalidJSON",
			id:         service.ID,
			json:       invalidJSON,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "EmptyJSON",
			id:         service.ID,
			json:       []byte("{}"),
			wantStatus: http.StatusOK,
		},
		{
			name:       "ValidJSON",
			id:         service.ID,
			json:       validJSON,
			wantStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req, _ := http.NewRequest("PUT", fmt.Sprintf("/services/%s", testCase.id), bytes.NewBuffer(testCase.json))
			req.Header.Set("Authorization", fmt.Sprintf("KeyPair %s:%s", rootSA.KeyID, rootSA.KeySecret))

			resp := helpers.DoRequest(t, req, app.GetRouter())
			if resp.Code != testCase.wantStatus {
				t.Errorf("HTTP Status = %v, want %v", resp.Code, testCase.wantStatus)
			}
		})
	}
}

func TestServicesUpdateHandler_servicePersisted(t *testing.T) {
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
		t.Fatalf("Create() returned error = %+v", service)
	}

	// Changes the data for testing
	service.Name = "Another name"
	service.PermissionName = "AnotherPermission"
	service.AMURL = "http://localhost:4444/am"

	reqJSON, err := json.Marshal(service)
	if err != nil {
		t.Fatalf("Marshal() returned error: %v", err)
	}
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/services/%s", service.ID), bytes.NewBuffer(reqJSON))
	req.Header.Set("Authorization", fmt.Sprintf("KeyPair %s:%s", rootSA.KeyID, rootSA.KeySecret))

	app := helpers.GetApp(t)
	helpers.DoRequest(t, req, app.GetRouter())

	services, err := servicesUC.List()
	if err != nil {
		t.Fatalf("List() returned error =  %v", err)
	}

	got := services[0]
	want := service
	// Fields set during Service update
	want.CreatedAt = got.CreatedAt
	want.UpdatedAt = got.UpdatedAt

	if diff := pretty.Compare(got, want); diff != "" {
		t.Errorf("%s: Services diff: (-got +want)\n%s", got, diff)
	}
}
