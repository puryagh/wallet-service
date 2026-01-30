DB_URL ?= postgresql://noghrestan:noghrestan_postgres_pass@localhost:6633/noghrestan?sslmode=disable
PG_CONTAINER ?= noghrestan-postgres
PG_USER ?= noghrestan
APPLICATION_NAME ?= wallet-service
ENV_DIR := ./env
CONSUL_HTTP_ADDR := localhost:8501

# If APPLICATION_NAME is exported in the environment but empty, ensure a sensible default
ifeq ($(strip $(APPLICATION_NAME)),)
APPLICATION_NAME := wallet-service
endif

.PHONY: help
help: ## Show this help message
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}'

.PHONY: proto
proto: ## Generate protobuf files, gRPC gateway, and OpenAPI docs (JSON + YAML)
	rm -f internal/abstract/pb/*.go
	mkdir -p internal/abstract/pb
	rm -f docs/*.swagger.json
	rm -f docs/*.yaml
	mkdir -p docs
	protoc --proto_path=proto --proto_path=/usr/include --go_out=internal/abstract/pb --go_opt=paths=source_relative \
	--go-grpc_out=internal/abstract/pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=internal/abstract/pb --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=docs --openapiv2_opt=logtostderr=true,json_names_for_fields=false \
	proto/*.proto
	@echo "Converting JSON to YAML..."
	@for file in docs/*.swagger.json; do \
		if [ -f "$$file" ]; then \
			base=$$(basename "$$file" .swagger.json); \
			yq eval -P "$$file" > "docs/$$base.yaml" 2>/dev/null || \
			python3 -c "import json, yaml, sys; yaml.dump(json.load(open('$$file')), sys.stdout, default_flow_style=False)" > "docs/$$base.yaml" 2>/dev/null || \
			echo "Warning: Could not convert $$file to YAML. Install yq or python3 with yaml module."; \
		fi \
	done

.PHONY: proto-yaml
proto-yaml: ## Convert existing OpenAPI JSON files to YAML format
	@echo "Converting existing JSON files to YAML..."
	@for file in docs/*.swagger.json; do \
		if [ -f "$$file" ]; then \
			base=$$(basename "$$file" .swagger.json); \
			yq eval -P "$$file" > "docs/$$base.yaml" 2>/dev/null || \
			python3 -c "import json, yaml, sys; yaml.dump(json.load(open('$$file')), sys.stdout, default_flow_style=False)" > "docs/$$base.yaml" 2>/dev/null || \
			echo "Warning: Could not convert $$file to YAML. Install yq or python3 with yaml module."; \
		fi \
	done

.PHONY: install-proto-deps
install-proto-deps: ## Install protobuf and gRPC-Gateway dependencies
	go env -w GOPROXY=https://goproxy.io,direct
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest

.PHONY: install-yaml-tools
install-yaml-tools: ## Install tools for YAML conversion (yq and/or PyYAML)
	@echo "Installing YAML conversion tools..."
	@which yq > /dev/null 2>&1 || (echo "Installing yq..." && \
		(brew install yq 2>/dev/null || \
		go env -w GOPROXY=https://goproxy.io,direct \
		go install github.com/mikefarah/yq/v4@latest 2>/dev/null || \
		echo "Please install yq manually: https://github.com/mikefarah/yq"))
	@python3 -c "import yaml" 2>/dev/null || \
		(echo "Installing Python YAML module..." && pip3 install PyYAML 2>/dev/null || \
		 echo "Please install PyYAML: pip3 install PyYAML")

.PHONY: createdb
createdb: ## Create expected database directly in the PostgreSQL container
	docker exec -it "$(PG_CONTAINER)" createdb --username="$(PG_USER)" --owner="$(PG_USER)" "$(PG_USER)"

.PHONY: dropdb
dropdb: ## Drop expected database directly in the PostgreSQL container
	docker exec -it "$(PG_CONTAINER)" dropdb --username="$(PG_USER)" "$(PG_USER)"

.PHONY: migrateup
migrateup: ## Apply all up migrations to the database from the migrations folder on ./internal/infra/db/migrations
	migrate -path internal/infra/db/postgres/migrations -database "$(DB_URL)" -verbose up

.PHONY: migratedown
migratedown: ## Apply all down migrations to the database from the migrations folder on ./internal/infra/db/migrations
	migrate -path internal/infra/db/postgres/migrations -database "$(DB_URL)" -verbose down
.PHONY: dbdump
dbdump: ## Dump the database to a SQL file at internal/infra/db/dump.sql
	docker exec -i "$(PG_CONTAINER)" psql -U "$(PG_USER)" -d "$(PG_USER)" < internal/infra/db/dump.sql

.PHONY: dbbackup
dbbackup: ## Backup the database to a SQL file at dbbackup/backup.sql
	mkdir -p dbbackup
	docker exec -t "$(PG_CONTAINER)" pg_dumpall -c -U "$(PG_USER)" > dbbackup/backup.sql

.PHONY: dbrestore
dbrestore: ## Restore the database from a SQL file at dbbackup/backup.sql
	docker exec -i "$(PG_CONTAINER)" psql -U "$(PG_USER)" -d "$(PG_USER)" < dbbackup/backup.sql

.PHONY: install-sqlc
install-sqlc: ## Install sqlc tool for generating type-safe database code
	go env -w GOPROXY=https://goproxy.io,direct
	go install github.com/kyleconroy/sqlc/cmd/sqlc@latest

.PHONY: sqlc
sqlc: ## Generate database query code using sqlc based on the configuration file at ./internal/infra/db/sqlc.yaml
	@sqlc generate -f ./internal/infra/db/postgres/sqlc.yaml

.PHONY: install-dbml2sql
install-dbml2sql: ## Install dbml2sql tool for converting DBML to SQL
	sudo npm install -g dbml2sql@latest

.PHONY: schema
schema: ## Generate SQL schema from DBML file located at internal/infra/db/diagram.dbml and output to internal/infra/db/schema.sql
	@echo "Generating SQL schema from DBML..."
	@dbml2sql internal/infra/db/postgres/database.dbml --postgres --out-file internal/infra/db/postgres/schema.sql
	@echo "Copying generated schema to postgres init directory..."
	@cp internal/infra/db/postgres/schema.sql ./postgres/schemas/schema.sql
	@echo "Copying database diagram to postgres database design directory..."
	@cp internal/infra/db/postgres/database.dbml ./docs/database.dbml
	@cp internal/infra/db/postgres/database.dbml ./postgres/design/database.dbml
	@cp internal/infra/db/postgres/database.dbml ../../noghrestan-stack/postgres/schemas/$(APPLICATION_NAME).dbml
	@cp internal/infra/db/postgres/schema.sql ../../noghrestan-stack/postgres/schemas/$(APPLICATION_NAME).sql
	@echo "Schema generation complete."

.PHONY: compose-dev-up
compose-dev-up: ## Start development containers
	docker compose -f docker-compose.dev.yaml --env-file $(ENV_DIR)/stack.dev.env up -d
	
.PHONY: compose-dev-down
compose-dev-down: ## Stop development containers
	docker compose -f docker-compose.dev.yaml --env-file $(ENV_DIR)/stack.dev.env down

.PHONY: compose-dev-down-volumes
compose-dev-down-volumes: ## Stop development containers and remove volumes
	docker compose -f docker-compose.dev.yaml --env-file $(ENV_DIR)/stack.dev.env down -v

.PHONY: audit
audit: ## Run gosec security analysis on the codebase excluding vendor directory, it requires gosec to be installed
	gosec ./...

.PHONY: heap-prof
heap-prof: ## Start a heap profile server on localhost:6060 and open pprof in the browser at localhost:6161
	go run ./cmd/$(APPLICATION_NAME)/main.go & \
		sleep 2 &&
	go tool pprof -http localhost:6161 http://localhost:6060/debug/pprof/heap

.PHONY: docker-build
docker-build: ## Build the Docker image for the application using BuildKit and SSH agent forwarding
	DOCKER_BUILDKIT=1 docker build -t $(APPLICATION_NAME) .

.PHONY: run
run: ## Run the application
	go run ./cmd/$(APPLICATION_NAME)/main.go

.PHONY: tidy
tidy: ## Tidy up Go module dependencies
	go env -w GOPROXY=https://goproxy.io,direct GOPRIVATE=github.com/liveutil/*
	go mod tidy

.PHONY: golib
golib: ## Install or update the liveutil/go-lib package that is private repository
	go env -w GOPROXY=https://goproxy.io,direct GOPRIVATE=github.com/liveutil/* 
	go get -u github.com/liveutil/go-lib

.PHONY: test
test: ## Run all tests in the test directory with verbose output
	go test -v ./test/...

.PHONY: test-api
test-api: ## Instructions to start servers for API testing
	@echo "Starting servers for API testing..."
	@echo "Make sure to run 'make run' in another terminal first"
	@echo "Then run: make test"

.PHONY: openapi-yaml
openapi-yaml: ## Display the locations of the generated OpenAPI spec files in YAML and JSON formats
	@echo "OpenAPI 3.0 spec is available at: docs/$(APPLICATION_NAME)_openapi.yaml"
	@echo "Swagger 2.0 spec is available at: docs/$(APPLICATION_NAME).swagger.json"

.PHONY: validate-openapi
validate-openapi: ## Validate the generated OpenAPI specification files
	@APP_NAME="$(strip $(if $(APP),$(APP),$(APPLICATION_NAME)))"; \
	echo "Using APPLICATION_NAME: $$APP_NAME"; \
	if [ -z "$$APP_NAME" ]; then echo "APPLICATION_NAME is not set"; exit 1; fi; \
	bash "$(CURDIR)/scripts/validate-openapi.sh" "$$APP_NAME"

.PHONY: init-db
init-db: ## Initialize the 'noghrestan' database in the Postgres container
	@echo "initializing the 'noghrestan' database in the Postgres container..."
	@echo "Input directory: postgres/schemas"
	@echo ""
	for file in postgres/schemas/*.sql; do \
		base=$$(basename $$file .sql); \
		docker exec -it noghrestan-postgres bash -c "psql -h localhost -p 6633 -U postgres -d noghrestan -f /docker-entrypoint-initdb.d/schemas/$$base.sql"; \
	done
	@echo "'noghrestan' database initialization completed."