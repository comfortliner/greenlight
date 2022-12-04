# ==========================================================================================================
# HELPERS
# ==========================================================================================================

# Include variables from the .envrc file
include .envrc

## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^//'


# ==========================================================================================================
# DEVELOPMENT
# ==========================================================================================================

## docker/mssql: run the mssql database docker container
.PHONY: docker/mssql
docker/mssql:
	docker run -d --name mssqlserver -e "ACCEPT_EULA=Y" -e "SA_PASSWORD=Pa55w0rd" -p 1433:1433 mcr.microsoft.com/mssql/server:2017-latest

## db/docker/createdb: create the 'greenlight' database
.PHONY: db/docker/createdb
db/docker/createdb:
	docker exec -it mssqlserver /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P Pa55w0rd -Q "CREATE DATABASE greenlight;"

## db/docker/dropdb: drop the 'greenlight' database
.PHONY: db/docker/dropdb
db/docker/dropdb:
	docker exec -it mssqlserver /opt/mssql-tools/bin/sqlcmd -S localhost -U sa -P Pa55w0rd -Q "DROP DATABASE greenlight;"

## db/migrations/new name=$1: create a new database migration
.PHONY: db/migrations/new
db/migrations/new:
	@echo 'Creating migration files for ${name}.'
	migrate create -seq -ext=.sql -dir=./migrations ${name}

## db/migrations/up: apply all up database migrations
.PHONY: db/migrations/up
db/migrations/up:
	@echo 'Running up migrations.'
	migrate -path migrations -database ${GREENLIGHT_DB_DSN} -verbose up

## db/migrations/up1: apply the next up database migration
.PHONY: db/migrations/up1
db/migrations/up1:
	migrate -path migrations -database ${GREENLIGHT_DB_DSN} -verbose up 1

## db/migrations/down: apply all down database migrations
.PHONY: db/migrations/down
db/migrations/down:
	@echo 'Running down migrations.'
	migrate -path migrations -database ${GREENLIGHT_DB_DSN} -verbose down

## db/migrations/down1: apply the further down database migration
.PHONY: db/migrations/down1 
db/migrations/down1:
	migrate -path migrations -database ${GREENLIGHT_DB_DSN} -verbose down 1


# ==========================================================================================================
# QUALITY CONTROL
# ==========================================================================================================

## qc/audit: tidy and vendor dependencies and format, vet and test all code
.PHONY: qc/audit
qc/audit: qc/vendor
	@echo 'Formatting code.'
	go fmt ./...
	@echo 'Vetting code.'
	go vet ./...
	staticcheck ./...
	@echo 'Running tests.'
	go test -race -vet=off ./...

## qc/vendor: tidy and vendor dependencies
.PHONY: qc/vendor
qc/vendor:
	@echo 'Tidying and verifying module dependencies.'
	go mod tidy
	go mod verify
	@echo 'Vendoring dependencies.'
	go mod vendor

# ==========================================================================================================
# RUN
# ==========================================================================================================

## api/test: test the cmd/api application
.PHONY: api/test
api/test:
	go test -v -cover ./...

## api/run: run the cmd/api application
.PHONY: api/run
api/run:
	@go run ./cmd/api/* -db-dsn=${GREENLIGHT_DB_DSN} -cors-trusted-origins="http://localhost:9000 http://localhost:9001"

# ==========================================================================================================
# BUILD
# ==========================================================================================================

## api/build: build the cmd/api application
.PHONY: api/build
api/build:
	@echo 'Building cmd/api for a macos/amd64  machine.'
	GOOS=darwin GOARCH=amd64 go build -o=./bin/darwin/api ./cmd/api
	@echo 'Building cmd/api for a windows/amd64 machine.'
	GOOS=windows GOARCH=amd64 go build -o=./bin/windows/api.exe ./cmd/api
