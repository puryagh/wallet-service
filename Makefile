DB_URL ?= postgresql://liveutil:liveutil_postgres_pass@localhost:6633/liveutil?sslmode=disable
PG_CONTAINER ?= liveutil-postgres
PG_USER ?= liveutil
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
	@cp internal/infra/db/postgres/database.dbml ../../liveutil-stack/postgres/schemas/$(APPLICATION_NAME).dbml
	@cp internal/infra/db/postgres/schema.sql ../../liveutil-stack/postgres/schemas/$(APPLICATION_NAME).sql
	@echo "Schema generation complete."

.PHONY: compose-tb-build
compose-tb-build: ## Build custom TigerBeetle image with healthcheck support
	@echo "Building custom TigerBeetle image..."
	docker build -t tigerbeetle-custom:latest ./tigerbeetle
	@echo "Custom TigerBeetle image built successfully."

.PHONY: compose-tb-format
compose-tb-format: ## Format TigerBeetle data files for the cluster (required before first start)
	@echo "Formatting TigerBeetle data files..."
	docker compose -f docker-compose.tb.yaml --env-file $(ENV_DIR)/tigerbittle.env run --rm tigerbeetle-0 format --cluster=0 --replica=0 --replica-count=3 /data/0_0.tigerbeetle
	docker compose -f docker-compose.tb.yaml --env-file $(ENV_DIR)/tigerbittle.env run --rm tigerbeetle-1 format --cluster=0 --replica=1 --replica-count=3 /data/0_1.tigerbeetle
	docker compose -f docker-compose.tb.yaml --env-file $(ENV_DIR)/tigerbittle.env run --rm tigerbeetle-2 format --cluster=0 --replica=2 --replica-count=3 /data/0_2.tigerbeetle
	@echo "TigerBeetle data files formatted successfully."

.PHONY: compose-tb-up
compose-tb-up: ## Start development TigerBittle cluster containers
	docker compose -f docker-compose.tb.yaml --env-file $(ENV_DIR)/tigerbittle.env up -d

.PHONY: compose-tb-down
compose-tb-down: ## Stop development TigerBittle cluster containers
	docker compose -f docker-compose.tb.yaml --env-file $(ENV_DIR)/tigerbittle.env down

.PHONY: compose-tb-down-volumes
compose-tb-down-volumes: ## Stop development TigerBittle cluster containers and remove volumes
	docker compose -f docker-compose.tb.yaml --env-file $(ENV_DIR)/tigerbittle.env down -v

.PHONY: compose-tb-init
compose-tb-init: ## Initialize TigerBeetle cluster from scratch (build, format and start)
	compose-tb-build compose-tb-down-volumes compose-tb-format compose-tb-up 

.PHONY: compose-tb-status
compose-tb-status: ## Show TigerBeetle cluster containers health status
	@echo "=== TigerBeetle Cluster Status ==="
	@docker ps --filter "name=tigerbeetle" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" 2>/dev/null || echo "No TigerBeetle containers running"
	@echo ""
	@echo "=== Detailed Health Information ==="
	@for container in tigerbeetle-0 tigerbeetle-1 tigerbeetle-2; do \
		if docker ps --filter "name=$$container" --format "{{.Names}}" 2>/dev/null | grep -q "$$container"; then \
			echo "--- $$container ---"; \
			docker inspect $$container --format='Status: {{.State.Health.Status}}' 2>/dev/null || echo "Status: No healthcheck configured"; \
			docker inspect $$container --format='Last Check: {{if .State.Health.Log}}{{(index .State.Health.Log 0).Output}}{{else}}N/A{{end}}' 2>/dev/null | head -1; \
			echo ""; \
		fi \
	done

.PHONY: compose-tb-logs
compose-tb-logs: ## Show TigerBeetle cluster logs (use CONTAINER=tigerbeetle-0 for specific container)
	@if [ -z "$(CONTAINER)" ]; then \
		echo "=== Showing logs for all TigerBeetle containers ==="; \
		docker compose -f docker-compose.tb.yaml --env-file $(ENV_DIR)/tigerbittle.env logs --tail=50 -f; \
	else \
		echo "=== Showing logs for $(CONTAINER) ==="; \
		docker logs -f $(CONTAINER); \
	fi

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
init-db: ## Initialize the 'liveutil' database in the Postgres container
	@echo "initializing the 'liveutil' database in the Postgres container..."
	@echo "Input directory: postgres/schemas"
	@echo ""
	for file in postgres/schemas/*.sql; do \
		base=$$(basename $$file .sql); \
		docker exec -it liveutil-postgres bash -c "psql -h localhost -p 6633 -U postgres -d liveutil -f /docker-entrypoint-initdb.d/schemas/$$base.sql"; \
	done
	@echo "'liveutil' database initialization completed."