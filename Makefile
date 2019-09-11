testable_packages=$(shell go list ./... | egrep -v 'constants|mocks|testing')
project=$(shell basename $(PWD))
project_test=${project}-test
pg_dep=$(project)_postgres_1
test_packages=`find . -type d -name "docker_data" -prune -o \
							-type f -name "*.go" ! \( -path "*vendor*" \) -print \
							| sed -En "s/([^\.])\/.*/\1/p" | uniq`
database=postgres://postgres:$(project)@localhost:8432/$(project)?sslmode=disable
database_test=postgres://postgres:$(project)@localhost:8432/$(project_test)?sslmode=disable
platform=darwin

export GO111MODULE=on

setup: setup-project setup-deps

setup-project:
	@go mod download

setup-deps:
	@make deps
	@make migrate

# run this if you don't have migrate
setup-migrate:
	@curl -L https://github.com/golang-migrate/migrate/releases/download/v4.4.0/migrate.$(platform)-amd64.tar.gz | tar xvz
	@mv migrate.$(platform)-amd64 /usr/local/bin/migrate

deps:
	@mkdir -p docker_data && docker-compose up -d postgres
	@until docker exec $(pg_dep) pg_isready; do echo 'Waiting Postgres...' && sleep 1; done
	@docker exec $(pg_dep) createuser -s -U postgres $(project) 2>/dev/null || true
	@docker exec $(pg_dep) createdb -U $(project) $(project) 2>/dev/null || true

deps-test:
	@mkdir -p docker_data && docker-compose up -d postgres
	@until docker exec $(pg_dep) pg_isready; do echo 'Waiting Postgres...' && sleep 1; done
	@docker exec $(pg_dep) createuser -s -U postgres $(project) 2>/dev/null || true
	@docker exec $(pg_dep) createdb -U $(project) $(project_test) 2>/dev/null || true
	@make migrate-test

stop-deps:
	@docker-compose down

stop-deps-test:
	@make drop-test
	@make stop-deps

build:
	@mkdir -p bin && go build -o ./bin/$(project) .

build-docker:
	@docker build -t $(project) .

run:
	@reflex -c reflex.conf -- sh -c ./Will.IAM start-api

migrate:
	@migrate -path migrations -database ${database} up

migrate-test:
	@migrate -path migrations -database ${database_test} up

drop:
	@migrate -path migrations -database ${database} drop

drop-test:
	@migrate -path migrations -database ${database_test} drop

test:
	@make deps-test
	@make test-fast
	@make stop-deps-test

test-fast:
	@make migrate-test
	@make unit
	@make integration
	@make drop-test

unit:
	@echo "Unit Tests"
	@go test ${testable_packages} -tags=unit -coverprofile unit.coverprofile -v
	@make gather-unit-profiles

integration:
	@echo "Integration Tests"
	@ret=0 && for pkg in ${testable_packages}; do \
		echo $$pkg; \
		go test $$pkg -tags=integration -coverprofile integration.coverprofile -v; \
		test $$? -eq 0 || ret=1; \
	done; exit $$ret
	@make gather-integration-profiles

gather-unit-profiles:
	@mkdir -p _build
	@echo "mode: count" > _build/coverage-unit.out
	@bash -c 'for f in $$(find . -type d -name "docker_data" -prune -o -type f \
		-name "*.coverprofile" -print); \
		do tail -n +2 $$f >> _build/coverage-unit.out; done'
	@find . -type d -name "docker_data" -prune -o \
		-name "*.coverprofile" -exec rm {} +

gather-integration-profiles:
	@mkdir -p _build
	@echo "mode: count" > _build/coverage-integration.out
	@bash -c 'for f in $$(find . -type d -name "docker_data" -prune -o -type f \
		-name "*.coverprofile" -print); \
		do tail -n +2 $$f >> _build/coverage-integration.out; done'
	@find . -type d -name "docker_data" -prune -o \
		-name "*.coverprofile" -exec rm {} +
