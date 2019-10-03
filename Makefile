testable_packages=$(shell go list ./... | egrep -v 'constants|mocks|testing')
project=$(shell basename $(PWD))
test_db_name=${project}-test
pg_docker_image=$(project)_postgres_1
db_url=postgres://postgres:$(project)@localhost:8432/$(project)?sslmode=disable
db_test_url=postgres://postgres:$(project)@localhost:8432/$(test_db_name)?sslmode=disable
uname_S=$(shell uname -s)

# TravisCI runs in Linux instances, while developers run in MacOS machines
ifeq ($(uname_S), Darwin)
  platform := darwin
else
  platform := linux
endif

export GO111MODULE=on

.PHONY: all
all: setup-migrate download-mod db-setup

.PHONY: all-ci
all-ci: setup-migrate download-mod db-setup-test

.PHONY: build
build:
	@mkdir -p bin && go build -o ./bin/$(project) .

.PHONY: build-docker
build-docker:
	@docker build -t $(project) .

.PHONY: run
run:
	@reflex -c reflex.conf -- sh -c ./bin/Will.IAM start-api

.PHONY: test
test: db-setup-test test-unit test-integration db-stop-test

# Installs the golang-migrate dependency if its not already installed.
.PHONY: setup-migrate
setup-migrate:
	@echo "Platform: ${platform}"
ifeq ($(shell command -v migrate),)
	@echo "Installing migrate..."
	@curl -L https://github.com/golang-migrate/migrate/releases/download/v4.4.0/migrate.$(platform)-amd64.tar.gz | tar xvz
	@mv migrate.$(platform)-amd64 $(GOPATH)/bin/migrate
	@echo "Done"
else
	@echo "migrate is already installed. Skipping..."
endif

.PHONY: download-mod
download-mod:
	@go mod download

.PHONY: compose-down
compose-down:
	@docker-compose down

.PHONY: db-setup
db-setup: db-up db-create-user db-create db-migrate

.PHONY: db-setup-test
db-setup-test: db-up db-create-user db-create-test db-migrate-test

.PHONY: db-up
db-up:
	@mkdir -p docker_data && docker-compose up -d postgres
	@until docker exec $(pg_docker_image) pg_isready; do echo 'Waiting Postgres...' && sleep 1; done
	@sleep 2

.PHONY: db-create-user
db-create-user:
	@docker exec $(pg_docker_image) createuser -s -U postgres $(project) 2>/dev/null || true

.PHONY: db-create
db-create:
	@docker exec $(pg_docker_image) createdb -U $(project) $(project) 2>/dev/null || true

.PHONY: db-create-test
db-create-test:
	@docker exec $(pg_docker_image) createdb -U $(project) $(test_db_name) 2>/dev/null || true
	@sleep 2

.PHONY: db-stop-test
db-stop-test: db-drop-test compose-down

.PHONY: db-drop-test
db-drop-test:
	@migrate -path migrations -database ${db_test_url} drop

.PHONY: db-migrate
db-migrate:
	@migrate -path migrations -database ${db_url} up

.PHONY: db-migrate-test
db-migrate-test:
	@migrate -path migrations -database ${db_test_url} up

.PHONY: db-drop
db-drop:
	@migrate -path migrations -database ${db_url} drop

.PHONY: test-unit
test-unit:
	@echo "Unit Tests"
	@go test ${testable_packages} -tags=unit -coverprofile unit.coverprofile -v
	@make gather-unit-profiles

.PHONY: test-integration
test-integration:
	@echo "Integration Tests"
	@ret=0 && for pkg in ${testable_packages}; do \
		echo $$pkg; \
		go test $$pkg -tags=integration -coverprofile integration.coverprofile -v; \
		test $$? -eq 0 || ret=1; \
	done; exit $$ret
	@make gather-integration-profiles

.PHONY: test-ci
test-ci:
	@echo "Unit Tests - START"
	@go test ${testable_packages} -tags=unit -covermode=count -coverprofile=coverage.out -v -p 1
	@echo "Unit Tests - DONE"
	@echo "Integration Tests - START"
	@go test ${testable_packages} -tags=integration -covermode=count -coverprofile=coverage.out -v -p 1
	@echo "Integration Tests - DONE"
	@goveralls -coverprofile=coverage.out -service=travis-ci -repotoken ${COVERALLS_TOKEN}

.PHONY: gather-unit-profiles
gather-unit-profiles:
	@mkdir -p _build
	@echo "mode: count" > _build/coverage-unit.out
	@bash -c 'for f in $$(find . -type d -name "docker_data" -prune -o -type f \
		-name "*.coverprofile" -print); \
		do tail -n +2 $$f >> _build/coverage-unit.out; done'
	@find . -type d -name "docker_data" -prune -o \
		-name "*.coverprofile" -exec rm {} +

.PHONY: gather-integration-profiles
gather-integration-profiles:
	@mkdir -p _build
	@echo "mode: count" > _build/coverage-integration.out
	@bash -c 'for f in $$(find . -type d -name "docker_data" -prune -o -type f \
		-name "*.coverprofile" -print); \
		do tail -n +2 $$f >> _build/coverage-integration.out; done'
	@find . -type d -name "docker_data" -prune -o \
		-name "*.coverprofile" -exec rm {} +

