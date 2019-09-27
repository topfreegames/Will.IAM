package client_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	william "github.com/topfreegames/Will.IAM/pkg/client"
)

func TestGenerateNil(t *testing.T) {
	f := william.Generate("RL", "Action")
	assert.NotNil(t, f)

	r, err := http.NewRequest(http.MethodGet, "", nil)
	assert.NoError(t, err)
	assert.Equal(t, "", f(r))
}

func TestGenerateInfoNil(t *testing.T) {
	f := william.GenerateInfo()
	assert.NotNil(t, f)

	r, err := http.NewRequest(http.MethodGet, "", nil)
	assert.NoError(t, err)
	assert.Equal(t, "", f(r))
}
