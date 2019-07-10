// +build integration

package usecases_test

import (
	"testing"

	"github.com/topfreegames/Will.IAM/models"
	"github.com/topfreegames/Will.IAM/repositories"
	helpers "github.com/topfreegames/Will.IAM/testing"
	"github.com/topfreegames/Will.IAM/usecases"
)

func TestPermissionsRequestsCreate(t *testing.T) {
	helpers.CleanupPG(t)
	saUC := helpers.GetServiceAccountsUseCase(t)
	saM := &models.ServiceAccount{
		Name:  "some name",
		Email: "test@domain.com",
	}
	if err := saUC.Create(saM); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	prsUC := helpers.GetPermissionsRequestsUseCase(t)
	pr := &models.PermissionRequest{
		ServiceAccountID:  saM.ID,
		Service:           "SomeService",
		OwnershipLevel:    models.OwnershipLevels.Lender,
		Action:            "Do",
		ResourceHierarchy: models.BuildResourceHierarchy("x::y"),
		Message:           "Please I need it",
	}
	if err := prsUC.Create(pr); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	storage := helpers.GetStorage(t)
	var prs []models.PermissionRequest
	storage.PG.DB.Query(&prs, "SELECT * FROM permissions_requests")
	if len(prs) != 1 {
		t.Errorf("Expected 1 permission request. Got %d", len(prs))
		return
	}
	if prs[0].State != models.PermissionRequestStates.Open {
		t.Errorf("Expected State to be open. Got %s", prs[0].State)
		return
	}
	if prs[0].ServiceAccountID != saM.ID {
		t.Errorf("Expected ServiceAccountID to be %s. Got %s", saM.ID, prs[0].ServiceAccountID)
		return
	}
	if prs[0].Service != "SomeService" {
		t.Errorf("Expected Service to be SomeService. Got %s", prs[0].Service)
		return
	}
	if prs[0].OwnershipLevel != models.OwnershipLevels.Lender {
		t.Errorf("Expected OwnershipLevel to be RL. Got %s", prs[0].OwnershipLevel)
		return
	}
	if prs[0].Action != "Do" {
		t.Errorf("Expected Action to be Do. Got %s", prs[0].Action)
		return
	}
	if prs[0].ResourceHierarchy != "x::y" {
		t.Errorf("Expected ResourceHierarchy to be x::y. Got %s", prs[0].ResourceHierarchy)
		return
	}
	if prs[0].Message != "Please I need it" {
		t.Errorf("Expected Message to be 'Please I need it'. Got %s", prs[0].Message)
		return
	}
	if prs[0].ModeratorServiceAccountID != "" {
		t.Errorf(
			"Expected ModeratorServiceAccountID to be empty. Got: %s",
			prs[0].ModeratorServiceAccountID,
		)
		return
	}
}

func TestPermissionsRequestsCreateDuplicate(t *testing.T) {
	helpers.CleanupPG(t)
	saUC := helpers.GetServiceAccountsUseCase(t)
	saM := &models.ServiceAccount{
		Name:  "some name",
		Email: "test@domain.com",
	}
	if err := saUC.Create(saM); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	prsUC := helpers.GetPermissionsRequestsUseCase(t)
	pr := &models.PermissionRequest{
		ServiceAccountID:  saM.ID,
		Service:           "SomeService",
		OwnershipLevel:    models.OwnershipLevels.Owner,
		Action:            "Do",
		ResourceHierarchy: models.BuildResourceHierarchy("x::y"),
		Message:           "Please I need it",
	}
	if err := prsUC.Create(pr); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	// ID == ""
	prD := &models.PermissionRequest{
		ServiceAccountID:  saM.ID,
		Service:           "SomeService",
		OwnershipLevel:    models.OwnershipLevels.Owner,
		Action:            "Do",
		ResourceHierarchy: models.BuildResourceHierarchy("x::y"),
		Message:           "Please I need it",
	}
	if err := prsUC.Create(prD); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	if prD.ID != "" {
		t.Error("Expected duplicate to have nil ID")
		return
	}
	storage := helpers.GetStorage(t)
	var prs []models.PermissionRequest
	storage.PG.DB.Query(&prs, "SELECT * FROM permissions_requests")
	if len(prs) != 1 {
		t.Errorf("Expected 1 permission request. Got %d", len(prs))
		return
	}
}

func TestPermissionsRequestsCreateWhenAlreadyHasPermission(t *testing.T) {
	helpers.CleanupPG(t)
	saUC := helpers.GetServiceAccountsUseCase(t)
	ps, err := models.BuildPermissions([]string{"SomeService::RO::Do::x::y"})
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	saM := &usecases.ServiceAccountWithNested{
		Name:               "some name",
		Email:              "test@domain.com",
		Permissions:        ps,
		AuthenticationType: models.AuthenticationTypes.OAuth2,
	}
	if err := saUC.CreateWithNested(saM); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	prsUC := helpers.GetPermissionsRequestsUseCase(t)
	pr := &models.PermissionRequest{
		ServiceAccountID:  saM.ID,
		Service:           "SomeService",
		OwnershipLevel:    models.OwnershipLevels.Owner,
		Action:            "Do",
		ResourceHierarchy: models.BuildResourceHierarchy("x::y"),
		Message:           "Please I need it",
	}
	if err := prsUC.Create(pr); err == nil || err.Error() != "user already has requested permission" {
		t.Errorf("Expected error 'user already has requested permission'")
		return
	}
	storage := helpers.GetStorage(t)
	var prs []models.PermissionRequest
	storage.PG.DB.Query(&prs, "SELECT * FROM permissions_requests")
	if len(prs) != 0 {
		t.Errorf("Expected 0 permission request. Got %d", len(prs))
		return
	}
}

func TestPermissionsRequestsListVisibleTo(t *testing.T) {
	helpers.CleanupPG(t)
	saUC := helpers.GetServiceAccountsUseCase(t)
	saM := &models.ServiceAccount{
		Name:  "some name",
		Email: "test@domain.com",
	}
	if err := saUC.Create(saM); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	prsUC := helpers.GetPermissionsRequestsUseCase(t)
	pr := &models.PermissionRequest{
		ServiceAccountID:  saM.ID,
		Service:           "SomeService",
		OwnershipLevel:    models.OwnershipLevels.Lender,
		Action:            "Do",
		ResourceHierarchy: models.BuildResourceHierarchy("x::y"),
		Message:           "Please I need it",
	}
	if err := prsUC.Create(pr); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	rootSA := helpers.CreateRootServiceAccount(t)
	prs, count, err := prsUC.ListOpenRequestsVisibleTo(&repositories.ListOptions{}, rootSA.ID)
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	if count != 1 {
		t.Errorf("Expected 1 visible open request. Got %d", count)
		return
	}
	if prs[0].State != models.PermissionRequestStates.Open {
		t.Errorf("Expected State to be open. Got %s", prs[0].State)
		return
	}
	if prs[0].ServiceAccountID != saM.ID {
		t.Errorf("Expected ServiceAccountID to be %s. Got %s", saM.ID, prs[0].ServiceAccountID)
		return
	}
	if prs[0].Service != "SomeService" {
		t.Errorf("Expected Service to be SomeService. Got %s", prs[0].Service)
		return
	}
	if prs[0].OwnershipLevel != models.OwnershipLevels.Lender {
		t.Errorf("Expected OwnershipLevel to be RL. Got %s", prs[0].OwnershipLevel)
		return
	}
	if prs[0].Action != "Do" {
		t.Errorf("Expected Action to be Do. Got %s", prs[0].Action)
		return
	}
	if prs[0].ResourceHierarchy != "x::y" {
		t.Errorf("Expected ResourceHierarchy to be x::y. Got %s", prs[0].ResourceHierarchy)
		return
	}
	if prs[0].Message != "Please I need it" {
		t.Errorf("Expected Message to be 'Please I need it'. Got %s", prs[0].Message)
		return
	}
}

func TestPermissionsRequestsListVisibleToWhenNonRoot(t *testing.T) {
	type testCase struct {
		saPs             []string
		reqs             []*models.PermissionRequest
		expectedReqsIdxs []int
	}
	testCases := []testCase{
		testCase{
			// same action, x::* <>- x::y, service = *
			saPs: []string{"*::RO::Do::x::*"},
			reqs: []*models.PermissionRequest{
				&models.PermissionRequest{
					Service:           "SomeService",
					OwnershipLevel:    models.OwnershipLevels.Lender,
					Action:            "Do",
					ResourceHierarchy: models.BuildResourceHierarchy("x::y"),
					Message:           "Please I need it",
				},
			},
			expectedReqsIdxs: []int{0},
		},
		testCase{
			// same as before, NOT OWNER
			saPs: []string{"*::RL::Do::x::*"},
			reqs: []*models.PermissionRequest{
				&models.PermissionRequest{
					Service:           "SomeService",
					OwnershipLevel:    models.OwnershipLevels.Lender,
					Action:            "Do",
					ResourceHierarchy: models.BuildResourceHierarchy("x::y"),
					Message:           "Please I need it",
				},
			},
			expectedReqsIdxs: []int{},
		},
		testCase{
			// same as first, with action = *
			saPs: []string{"*::RO::*::x::*"},
			reqs: []*models.PermissionRequest{
				&models.PermissionRequest{
					Service:           "SomeService",
					OwnershipLevel:    models.OwnershipLevels.Lender,
					Action:            "Do",
					ResourceHierarchy: models.BuildResourceHierarchy("x::y"),
					Message:           "Please I need it",
				},
			},
			expectedReqsIdxs: []int{0},
		},
		testCase{
			// XX != Do
			saPs: []string{"*::RO::XX::x::*"},
			reqs: []*models.PermissionRequest{
				&models.PermissionRequest{
					Service:           "SomeService",
					OwnershipLevel:    models.OwnershipLevels.Lender,
					Action:            "Do",
					ResourceHierarchy: models.BuildResourceHierarchy("x::y"),
					Message:           "Please I need it",
				},
			},
			expectedReqsIdxs: []int{},
		},
		testCase{
			saPs: []string{"*::RO::*::x::*"},
			reqs: []*models.PermissionRequest{
				&models.PermissionRequest{
					Service:           "SomeOther",
					OwnershipLevel:    models.OwnershipLevels.Lender,
					Action:            "YY",
					ResourceHierarchy: models.BuildResourceHierarchy("x::y"),
					Message:           "Please I need it",
				},
				&models.PermissionRequest{
					Service:           "SomeService",
					OwnershipLevel:    models.OwnershipLevels.Lender,
					Action:            "XX",
					ResourceHierarchy: models.BuildResourceHierarchy("x::y"),
					Message:           "Please I need it",
				},
			},
			expectedReqsIdxs: []int{0, 1},
		},
		testCase{
			saPs: []string{"*::RO::XX::x::*"},
			reqs: []*models.PermissionRequest{
				&models.PermissionRequest{
					Service:           "SomeOther",
					OwnershipLevel:    models.OwnershipLevels.Lender,
					Action:            "YY",
					ResourceHierarchy: models.BuildResourceHierarchy("x::y"),
					Message:           "Please I need it",
				},
				&models.PermissionRequest{
					Service:           "SomeService",
					OwnershipLevel:    models.OwnershipLevels.Lender,
					Action:            "XX",
					ResourceHierarchy: models.BuildResourceHierarchy("x::y"),
					Message:           "Please I need it",
				},
			},
			expectedReqsIdxs: []int{1},
		},
	}

	for i, tt := range testCases {
		helpers.CleanupPG(t)
		saUC := helpers.GetServiceAccountsUseCase(t)
		ps, err := models.BuildPermissions(tt.saPs)
		if err != nil {
			t.Errorf("Unexpected error: %s. Case: %d", err.Error(), i)
			return
		}
		sa := &usecases.ServiceAccountWithNested{
			Name:               "some name",
			Email:              "test@domain.com",
			Permissions:        ps,
			AuthenticationType: models.AuthenticationTypes.OAuth2,
		}
		if err := saUC.CreateWithNested(sa); err != nil {
			t.Errorf("Unexpected error: %s. Case: %d", err.Error(), i)
		}
		prsUC := helpers.GetPermissionsRequestsUseCase(t)
		saL := &models.ServiceAccount{
			Name:  "sa_lender",
			Email: "lender@domain.com",
		}
		if err := saUC.Create(saL); err != nil {
			t.Errorf("Unexpected error: %s", err.Error())
		}
		for _, pr := range tt.reqs {
			pr.ServiceAccountID = saL.ID
			if err := prsUC.Create(pr); err != nil {
				t.Errorf("Unexpected error: %s. Case: %d", err.Error(), i)
				return
			}
		}
		prs, count, err := prsUC.ListOpenRequestsVisibleTo(&repositories.ListOptions{}, sa.ID)
		if err != nil {
			t.Errorf("Unexpected error: %s. Case: %d", err.Error(), i)
			return
		}
		if count != int64(len(tt.expectedReqsIdxs)) {
			t.Errorf("Expected 1 visible open request. Got %d. Case: %d", count, i)
			return
		}
		for j, idx := range tt.expectedReqsIdxs {
			if tt.reqs[idx].ID != prs[j].ID {
				t.Errorf("Expected PermissionRequest.ID to be %s. Got %s. Case: %d", tt.reqs[idx].ID, prs[j].ID, i)
				return
			}
		}
	}
}

func TestPermissionsRequestsGrant(t *testing.T) {
	helpers.CleanupPG(t)
	saUC := helpers.GetServiceAccountsUseCase(t)
	saM := &models.ServiceAccount{
		Name:  "some name",
		Email: "test@domain.com",
	}
	if err := saUC.Create(saM); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	prsUC := helpers.GetPermissionsRequestsUseCase(t)
	pr := &models.PermissionRequest{
		ServiceAccountID:  saM.ID,
		Service:           "SomeService",
		OwnershipLevel:    models.OwnershipLevels.Lender,
		Action:            "Do",
		ResourceHierarchy: models.BuildResourceHierarchy("x::y"),
		Message:           "Please I need it",
	}
	if err := prsUC.Create(pr); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	sasUC := helpers.GetServiceAccountsUseCase(t)
	has, err := sasUC.HasPermissionString(saM.ID, pr.Permission().String())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	if has {
		t.Errorf("Expected saM to NOT have permission")
		return
	}
	rootSA := helpers.CreateRootServiceAccount(t)
	if err := prsUC.Grant(rootSA.ID, pr.ID); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	storage := helpers.GetStorage(t)
	var prs []models.PermissionRequest
	storage.PG.DB.Query(&prs, "SELECT * FROM permissions_requests")
	if len(prs) != 1 {
		t.Errorf("Expected 1 permission request. Got %d", len(prs))
		return
	}
	if prs[0].State != models.PermissionRequestStates.Granted {
		t.Errorf("Expected State to be granted. Got %s", prs[0].State)
		return
	}
	if prs[0].ServiceAccountID != saM.ID {
		t.Errorf("Expected ServiceAccountID to be %s. Got %s", saM.ID, prs[0].ServiceAccountID)
		return
	}
	if prs[0].Service != "SomeService" {
		t.Errorf("Expected Service to be SomeService. Got %s", prs[0].Service)
		return
	}
	if prs[0].OwnershipLevel != models.OwnershipLevels.Lender {
		t.Errorf("Expected OwnershipLevel to be RL. Got %s", prs[0].OwnershipLevel)
		return
	}
	if prs[0].Action != "Do" {
		t.Errorf("Expected Action to be Do. Got %s", prs[0].Action)
		return
	}
	if prs[0].ResourceHierarchy != "x::y" {
		t.Errorf("Expected ResourceHierarchy to be x::y. Got %s", prs[0].ResourceHierarchy)
		return
	}
	if prs[0].Message != "Please I need it" {
		t.Errorf("Expected Message to be 'Please I need it'. Got %s", prs[0].Message)
		return
	}
	if prs[0].ModeratorServiceAccountID != rootSA.ID {
		t.Errorf(
			"Expected ModeratorServiceAccountID to be rootSA.ID. Got: %s",
			prs[0].ModeratorServiceAccountID,
		)
		return
	}
	has, err = sasUC.HasPermissionString(saM.ID, pr.Permission().String())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	if !has {
		t.Errorf("Expected saM to have permission")
		return
	}
}

func TestPermissionsRequestsDeny(t *testing.T) {
	helpers.CleanupPG(t)
	saUC := helpers.GetServiceAccountsUseCase(t)
	saM := &models.ServiceAccount{
		Name:  "some name",
		Email: "test@domain.com",
	}
	if err := saUC.Create(saM); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	prsUC := helpers.GetPermissionsRequestsUseCase(t)
	pr := &models.PermissionRequest{
		ServiceAccountID:  saM.ID,
		Service:           "SomeService",
		OwnershipLevel:    models.OwnershipLevels.Lender,
		Action:            "Do",
		ResourceHierarchy: models.BuildResourceHierarchy("x::y"),
		Message:           "Please I need it",
	}
	if err := prsUC.Create(pr); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	rootSA := helpers.CreateRootServiceAccount(t)
	if err := prsUC.Deny(rootSA.ID, pr.ID); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	storage := helpers.GetStorage(t)
	var prs []models.PermissionRequest
	storage.PG.DB.Query(&prs, "SELECT * FROM permissions_requests")
	if len(prs) != 1 {
		t.Errorf("Expected 1 permission request. Got %d", len(prs))
		return
	}
	if prs[0].State != models.PermissionRequestStates.Denied {
		t.Errorf("Expected State to be denied. Got %s", prs[0].State)
		return
	}
	if prs[0].ServiceAccountID != saM.ID {
		t.Errorf("Expected ServiceAccountID to be %s. Got %s", saM.ID, prs[0].ServiceAccountID)
		return
	}
	if prs[0].Service != "SomeService" {
		t.Errorf("Expected Service to be SomeService. Got %s", prs[0].Service)
		return
	}
	if prs[0].OwnershipLevel != models.OwnershipLevels.Lender {
		t.Errorf("Expected OwnershipLevel to be RL. Got %s", prs[0].OwnershipLevel)
		return
	}
	if prs[0].Action != "Do" {
		t.Errorf("Expected Action to be Do. Got %s", prs[0].Action)
		return
	}
	if prs[0].ResourceHierarchy != "x::y" {
		t.Errorf("Expected ResourceHierarchy to be x::y. Got %s", prs[0].ResourceHierarchy)
		return
	}
	if prs[0].Message != "Please I need it" {
		t.Errorf("Expected Message to be 'Please I need it'. Got %s", prs[0].Message)
		return
	}
	if prs[0].ModeratorServiceAccountID != rootSA.ID {
		t.Errorf(
			"Expected ModeratorServiceAccountID to be rootSA.ID. Got: %s",
			prs[0].ModeratorServiceAccountID,
		)
		return
	}
	sasUC := helpers.GetServiceAccountsUseCase(t)
	has, err := sasUC.HasPermissionString(saM.ID, pr.Permission().String())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	if has {
		t.Errorf("Expected saM to NOT have permission")
		return
	}
}

func TestPermissionsRequestsGrantWhenPRIsNotOpen(t *testing.T) {
	helpers.CleanupPG(t)
	saUC := helpers.GetServiceAccountsUseCase(t)
	saM := &models.ServiceAccount{
		Name:  "some name",
		Email: "test@domain.com",
	}
	if err := saUC.Create(saM); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	prsUC := helpers.GetPermissionsRequestsUseCase(t)
	pr := &models.PermissionRequest{
		ServiceAccountID:  saM.ID,
		Service:           "SomeService",
		OwnershipLevel:    models.OwnershipLevels.Lender,
		Action:            "Do",
		ResourceHierarchy: models.BuildResourceHierarchy("x::y"),
		Message:           "Please I need it",
	}
	if err := prsUC.Create(pr); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	rootSA := helpers.CreateRootServiceAccount(t)
	if err := prsUC.Deny(rootSA.ID, pr.ID); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	if err := prsUC.Grant(rootSA.ID, pr.ID); err == nil || err.Error() != "permission request is closed" {
		t.Error("Expected error to be 'permission request is closed'")
		return
	}
	sasUC := helpers.GetServiceAccountsUseCase(t)
	has, err := sasUC.HasPermissionString(saM.ID, pr.Permission().String())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	if has {
		t.Errorf("Expected saM to NOT have permission")
		return
	}
}

func TestPermissionsRequestsDenyWhenPRIsNotOpen(t *testing.T) {
	helpers.CleanupPG(t)
	saUC := helpers.GetServiceAccountsUseCase(t)
	saM := &models.ServiceAccount{
		Name:  "some name",
		Email: "test@domain.com",
	}
	if err := saUC.Create(saM); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
	}
	prsUC := helpers.GetPermissionsRequestsUseCase(t)
	pr := &models.PermissionRequest{
		ServiceAccountID:  saM.ID,
		Service:           "SomeService",
		OwnershipLevel:    models.OwnershipLevels.Lender,
		Action:            "Do",
		ResourceHierarchy: models.BuildResourceHierarchy("x::y"),
		Message:           "Please I need it",
	}
	if err := prsUC.Create(pr); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	rootSA := helpers.CreateRootServiceAccount(t)
	if err := prsUC.Deny(rootSA.ID, pr.ID); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	if err := prsUC.Deny(rootSA.ID, pr.ID); err == nil || err.Error() != "permission request is closed" {
		t.Error("Expected error to be 'permission request is closed'")
		return
	}
	sasUC := helpers.GetServiceAccountsUseCase(t)
	has, err := sasUC.HasPermissionString(saM.ID, pr.Permission().String())
	if err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		return
	}
	if has {
		t.Errorf("Expected saM to NOT have permission")
		return
	}
}
