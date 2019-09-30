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

	invalidService := &models.Service{
		Name:                    "",
		PermissionName:          "SomeService",
		CreatorServiceAccountID: sa.ID,
		AMURL:                   "http://localhost:3333/am",
	}

	validJson, err := json.Marshal(service)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	invalidJson, err := json.Marshal(invalidService)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
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
			json:       invalidJson,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "ValidJSON",
			json:       validJson,
			wantStatus: http.StatusCreated,
		},
	}

	app := helpers.GetApp(t)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(testCase.json))
			req.Header.Set("Authorization", fmt.Sprintf("KeyPair %s:%s", rootSA.KeyID, rootSA.KeySecret))

			rec := helpers.DoRequest(t, req, app.GetRouter())
			if rec.Code != testCase.wantStatus {
				t.Fatalf("Expected %v got %v", testCase.wantStatus, rec.Code)
			}
		})
	}
}

func TestServicesCreateHandlerServicePersistence(t *testing.T) {
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

	reqJson, err := json.Marshal(service)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	req, _ := http.NewRequest("POST", "/services", bytes.NewBuffer(reqJson))
	req.Header.Set("Authorization", fmt.Sprintf("KeyPair %s:%s", rootSA.KeyID, rootSA.KeySecret))

	app := helpers.GetApp(t)
	helpers.DoRequest(t, req, app.GetRouter())

	servicesUC := helpers.GetServicesUseCase(t)
	services, err := servicesUC.List()
	if err != nil {
		t.Fatalf("Unable to list Services, error %v", err)
	}

	savedService := services[0]
	if savedService.Name != service.Name {
		t.Errorf("Expected name %v, got %v", service.Name, savedService.Name)
	}

	if savedService.PermissionName != service.PermissionName {
		t.Errorf("Expected permissionName %v, got %v", service.PermissionName, savedService.PermissionName)
	}

	if savedService.CreatorServiceAccountID != rootSA.ID {
		t.Errorf("Expected CreatorServiceAccountID %v, got %v", rootSA.ID, savedService.CreatorServiceAccountID)
	}

	if savedService.AMURL != service.AMURL {
		t.Errorf("Expected AMURL %v, got %v", service.AMURL, savedService.AMURL)
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

			rec := helpers.DoRequest(t, req, app.GetRouter())

			if rec.Code != testCase.wantStatus {
				t.Errorf("Want HTTP Status %v, got %v", testCase.wantStatus, rec.Code)
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
	err := servicesUC.Create(service)
	if err != nil {
		t.Fatalf("Error persisting Service = %+v", service)
	}

	service.Name = "Another name"
	validJson, err := json.Marshal(service)
	if err != nil {
		t.Fatalf("Error marshalling JSON: %v", err)
	}

	service.Name = ""
	invalidJson, err := json.Marshal(service)
	if err != nil {
		t.Fatalf("Error marshalling JSON: %v", err)
	}

	app := helpers.GetApp(t)

	testCases := []struct {
		name       string
		id         string
		json       []byte
		wantStatus int
	}{
		{
			name:       "EmptyJSON",
			id:         service.ID,
			json:       []byte("{}"),
			wantStatus: http.StatusUnprocessableEntity,
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
			json:       invalidJson,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "ValidJSON",
			id:         service.ID,
			json:       validJson,
			wantStatus: http.StatusOK,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			req, _ := http.NewRequest("PUT", fmt.Sprintf("/services/%s", testCase.id), bytes.NewBuffer(testCase.json))
			req.Header.Set("Authorization", fmt.Sprintf("KeyPair %s:%s", rootSA.KeyID, rootSA.KeySecret))

			rec := helpers.DoRequest(t, req, app.GetRouter())

			if rec.Code != testCase.wantStatus {
				t.Errorf("Want HTTP Status %v, got %v", testCase.wantStatus, rec.Code)
			}
		})
	}
}

func TestServicesUpdateHandlerServicePersistence(t *testing.T) {
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
		t.Fatalf("Error persisting Service = %+v", service)
	}

	// Changes the data for testing
	service.Name = "Another name"
	service.PermissionName = "AnotherPermission"
	service.AMURL = "http://localhost:4444/am"

	reqJson, err := json.Marshal(service)
	if err != nil {
		t.Fatalf("Unexpected error: %s", err.Error())
	}

	req, _ := http.NewRequest("PUT", fmt.Sprintf("/services/%s", service.ID), bytes.NewBuffer(reqJson))
	req.Header.Set("Authorization", fmt.Sprintf("KeyPair %s:%s", rootSA.KeyID, rootSA.KeySecret))

	app := helpers.GetApp(t)
	helpers.DoRequest(t, req, app.GetRouter())

	services, err := servicesUC.List()
	if err != nil {
		t.Fatalf("Unable to list Services, error %v", err)
	}

	savedService := services[0]
	if savedService.Name != "Another name" {
		t.Errorf("Expected name %v, got %v", "Another name", savedService.Name)
	}

	if savedService.PermissionName != "AnotherPermission" {
		t.Errorf("Expected permissionName %v, got %v", "AnotherPermission", savedService.PermissionName)
	}

	if savedService.CreatorServiceAccountID != rootSA.ID {
		t.Errorf("Expected CreatorServiceAccountID %v, got %v", rootSA.ID, savedService.CreatorServiceAccountID)
	}

	if savedService.AMURL != "http://localhost:4444/am" {
		t.Errorf("Expected AMURL %v, got %v", "http://localhost:4444/am", savedService.AMURL)
	}

}
