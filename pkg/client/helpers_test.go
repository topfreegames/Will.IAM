package client_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	william "github.com/topfreegames/Will.IAM/pkg/client"
)

func TestGenerateNil(t *testing.T) {
	f := william.Generate("RL", "Action")
	assert.NotNil(t, f)
	assert.Equal(t, "", f(nil))
}

func TestGenerateInfoNil(t *testing.T) {
	f := william.GenerateInfo()
	assert.NotNil(t, f)
	assert.Equal(t, "", f(nil))
}
