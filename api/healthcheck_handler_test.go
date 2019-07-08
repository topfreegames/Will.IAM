// +build integration

package api_test

import (
	"net/http"
	"testing"

	helpers "github.com/topfreegames/Will.IAM/testing"
)

func TestHealthcheckTrue(t *testing.T) {
	app := helpers.GetApp(t)
	req, _ := http.NewRequest("GET", "/healthcheck", nil)
	rec := helpers.DoRequest(t, req, app.GetRouter())
	if body := rec.Body.String(); body != `{"healthy": true}` {
		t.Errorf("Expected healthy. Got %s", body)
	}
}
