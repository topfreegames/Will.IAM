// +build unit

package api_test

import (
	"testing"

	"github.com/ghostec/Will.IAM/api"
)

func TestReplaceIdInPermission(t *testing.T) {
	m := map[string]string{
		"id": "some-id",
	}
	p := api.ReplaceRequestVarsInPermission(m, "Will.IAM::RL::EditRole::{id}")
	e := "Will.IAM::RL::EditRole::some-id"
	if p != e {
		t.Errorf("Expected permission to be %s. Got %s", e, p)
	}
}
