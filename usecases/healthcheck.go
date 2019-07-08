package usecases

import "github.com/topfreegames/Will.IAM/repositories"

// Healthcheck usecase
type Healthcheck interface {
	Do() error
}

type healthcheck struct {
	repo *repositories.All
}

func (h healthcheck) Do() error {
	return h.repo.Healthcheck.Do()
}

// NewHealthcheck ctor
func NewHealthcheck(repo *repositories.All) Healthcheck {
	return &healthcheck{repo: repo}
}
