package client_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	william "github.com/topfreegames/Will.IAM/pkg/client"
)

func TestWilliamListPermission(t *testing.T) {
	ctx := context.Background()
	williamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/service_accounts/with_permission", r.URL.Path)
		assert.Equal(t, "ServicesStatus::RL::Something::Game", r.URL.Query()["permission"][0])
		assert.Equal(t, "KeyPair CLIENT_ID:CLIENT_SECRET", r.Header.Get("Authorization"))

		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"count":0,"result":[]}`)
	}))
	defer williamServer.Close()

	wi := william.New(williamServer.URL, "ServicesStatus")
	wi.SetClient(williamServer.Client())
	wi.SetKeyPair("CLIENT_ID", "CLIENT_SECRET")

	list, err := wi.ListPermission(ctx, "RL", "Something", "Game")
	assert.NoError(t, err)
	assert.JSONEq(t, "{\"count\":0,\"result\":[]}\n", string(list))
}

func TestWilliamPermissionWithNewAuthToken(t *testing.T) {
	authToken := "Bearer ya29.Il-GB5VlfTHcaUr4V4T0KrLswMdE7ej0Df0Ry6ny9rZ27ygNJBVwYMuKa_gCAwN_zATbWbdQqZo7lctXsAMmbZvoofGh07yX81ZrQh0VXuRc7gircolwtVuKhkjhdaY7fP"
	authTokenNew := "Bearer ya29.IZ-EB5VlfTHcaUr4V4T0KrLswMdE7ej0Df0Ry6ny9rZ27ygNJBVwYMuKa_gCAwN_zATbWbdQqZo7lctXsAMmbZvoofGh07yX81ZrQh0VXuRc7gircolwtVuKhkjhdaY7fP"
	email := "fake@email.com"

	williamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "/permissions/has", r.URL.Path)
		assert.Equal(t, "ServicesStatus::RL::Action::*", r.URL.Query()["permission"][0])
		assert.Equal(t, authToken, r.Header.Get("Authorization"))

		w.Header().Set("x-email", email)
		w.Header().Set("x-access-token", authTokenNew)
		w.WriteHeader(http.StatusOK)
	}))
	defer williamServer.Close()

	wi := william.New(williamServer.URL, "ServicesStatus")
	wi.SetClient(williamServer.Client())
	wi.SetKeyPair("CLIENT_ID", "CLIENT_SECRET")

	req := httptest.NewRequest(http.MethodGet, "/action", nil)
	req.Header.Set("Authorization", authToken)

	rr := httptest.NewRecorder()
	wi.HandlerFunc(wi.Generate("RL", "Action"),
		func(w http.ResponseWriter, r *http.Request) {
			auth := william.Auth(r)
			assert.NotNil(t, auth)

			assert.Equal(t, email, auth.Email())
			assert.Equal(t, "", auth.Name())
			assert.Equal(t, authTokenNew, auth.Token())
			assert.Equal(t, http.StatusOK, auth.Code())
			assert.Equal(t, "ServicesStatus::RL::Action::*", auth.Permission())
			assert.Empty(t, auth.Body())
			w.WriteHeader(http.StatusOK)
		}).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, email, rr.Header().Get("x-email"))
	assert.Equal(t, authTokenNew, rr.Header().Get("x-access-token"))
}

func TestWilliamPermissionDissable(t *testing.T) {
	williamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Fail(t, "willian is enable")
		w.WriteHeader(http.StatusOK)
	}))
	defer williamServer.Close()

	wi := william.New(williamServer.URL, "ServicesStatus")
	wi.SetClient(williamServer.Client())
	wi.SetKeyPair("CLIENT_ID", "CLIENT_SECRET")
	wi.ByPass()

	req := httptest.NewRequest(http.MethodGet, "/action", nil)

	rr := httptest.NewRecorder()
	wi.HandlerFunc(wi.Generate("RL", "Action"),
		func(w http.ResponseWriter, r *http.Request) {
			auth := william.Auth(r)
			assert.Nil(t, auth)

			w.WriteHeader(http.StatusOK)
		}).ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "", rr.Header().Get("x-email"))
	assert.Equal(t, "", rr.Header().Get("x-access-token"))
}

func TestGetServiceName(t *testing.T) {
	wi := william.New("http://fakeurl.com", "ServicesStatus")

	assert.Equal(t, "ServicesStatus", wi.GetServiceName())
}

func TestAmListStatic(t *testing.T) {
	wi := william.New("http://fakeurl.com", "ServicesStatus")

	wi.AddAction("GetAction")
	wi.AddAction("PostAction")

	req := httptest.NewRequest(http.MethodGet, "/am", nil)
	rr := httptest.NewRecorder()
	wi.AmHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	compareJsonArrayPermissions(t, "[{\"prefix\":\"GetAction\",\"complete\":false},{\"prefix\":\"PostAction\",\"complete\":false}]\n", rr.Body.String())
}

func TestAmListStaticAction(t *testing.T) {
	wi := william.New("http://fakeurl.com", "ServicesStatus")

	wi.AddAction("GetAction")
	wi.AddAction("PostAction")

	req := httptest.NewRequest(http.MethodGet, "/am?prefix=GetAction::", nil)
	rr := httptest.NewRecorder()
	wi.AmHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	compareJsonArrayPermissions(t, "[{\"prefix\":\"GetAction::*\",\"complete\":true}]\n", rr.Body.String())
}

func TestAmListDynamicAction(t *testing.T) {
	wi := william.New("http://fakeurl.com", "ServicesStatus")

	wi.AddAction("GetAction")
	wi.AddActionFunc("PostAction", func(ctx context.Context, prefix string) []william.AmPermission {
		assert.Equal(t, "PostAction::", prefix)

		return []william.AmPermission{
			{
				Alias:    "Game Name",
				Prefix:   prefix + "gameid",
				Complete: true,
			},
		}
	})

	req := httptest.NewRequest(http.MethodGet, "/am?prefix=PostAction::", nil)
	rr := httptest.NewRecorder()
	wi.AmHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	compareJsonArrayPermissions(t, "[{\"prefix\":\"PostAction::*\",\"complete\":true},{\"alias\":\"Game Name\",\"prefix\":\"PostAction::gameid\",\"complete\":true}]\n", rr.Body.String())
}

func compareJsonArrayPermissions(t *testing.T, expected, actual string) bool {
	t.Helper()

	var expectedJSON, actualJSON []william.AmPermission

	if err := json.Unmarshal([]byte(expected), &expectedJSON); err != nil {
		return assert.Fail(t, fmt.Sprintf("Expected value ('%s') is not valid json.\nJSON parsing error: '%s'", expected, err.Error()))
	}

	if err := json.Unmarshal([]byte(actual), &actualJSON); err != nil {
		return assert.Fail(t, fmt.Sprintf("Input ('%s') needs to be valid json.\nJSON parsing error: '%s'", actual, err.Error()))
	}

	sort.Slice(expectedJSON, func(i, j int) bool { return expectedJSON[i].Prefix < expectedJSON[j].Prefix })
	sort.Slice(actualJSON, func(i, j int) bool { return actualJSON[i].Prefix < actualJSON[j].Prefix })

	return assert.Equal(t, expectedJSON, actualJSON)

}
