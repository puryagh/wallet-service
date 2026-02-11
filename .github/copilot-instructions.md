# Copilot Instructions for Wallet Service

## Overview
This repository contains the Wallet Service, a high-performance, reusable multi-asset wallet service with double-entry bookkeeping, powered by the TigerBeetle distributed database. The service ensures ACID compliance and high-performance transaction processing for financial applications.

## Architecture
The Wallet Service is built with the following key components:

- **TigerBeetle Cluster**: A distributed financial accounting database with 3 replicas for high availability. It ensures fault tolerance, high performance, and strict consistency for financial transactions.
  - **Key Files**:
    - `tigerbeetle/Dockerfile`: Custom Docker image for TigerBeetle with additional tools for health monitoring.
    - `docker-compose.tb.yaml`: Configuration for setting up the TigerBeetle cluster.
- **Go Backend**: The core service logic is implemented in Go.
  - **Key Files**:
    - `cmd/wallet-service/main.go`: Entry point for the wallet service.
    - `internal/infra/db/postgres`: Contains database-related code, including migrations, queries, and repository implementations.
    - `internal/config/config.go`: Configuration management for the service.
- **Protobuf Definitions**: Defines the gRPC APIs and data models.
  - **Key Files**:
    - `proto/rpc_wallet_service.proto`: Main gRPC service definition.
    - `internal/abstract/pb`: Generated protobuf and gRPC gateway files.

## Developer Workflows

### Building the Project
Use the `Makefile` to manage builds and other tasks. Key commands include:

- `make build`: Build the wallet service.
- `make schema`: Generate database schema and related files.
- `make run`: Start the wallet service.

### Running Tests
- Use `make test` to run the test suite.

### Debugging
- Use `make debug` to start the service in debug mode.
- Logs are available for debugging in the `logs/` directory.

### Database Migrations
- Migrations are stored in `internal/infra/db/postgres/migrations/`.
- Use `make migrate` to apply migrations.

### TigerBeetle Setup
- The TigerBeetle cluster is configured using `docker-compose.tb.yaml`.
- Use `make tb-up` to start the cluster.
- Use `make tb-down` to stop the cluster.

## Project-Specific Conventions
- **Double-Entry Bookkeeping**: All financial transactions must adhere to double-entry principles.
- **Protobuf Usage**: All APIs are defined using Protobuf in the `proto/` directory. Generated files are stored in `internal/abstract/pb/`.
- **Repository Pattern**: Database interactions are implemented using the repository pattern in `internal/infra/db/postgres/repository/`.

## External Dependencies
- **TigerBeetle**: A distributed financial accounting database.
- **PostgreSQL**: Used for persistent storage.
- **Docker**: For containerized development and deployment.

## Key Files and Directories
- `cmd/wallet-service/main.go`: Main entry point for the service.
- `internal/infra/db/postgres`: Database schema, queries, and repository implementations.
- `proto/`: Protobuf definitions for gRPC APIs.
- `tigerbeetle/`: Custom Docker image and configuration for TigerBeetle.

## Notes for AI Agents
- Focus on maintaining strict double-entry bookkeeping principles in all financial operations.
- Follow the repository pattern for database interactions.
- Ensure all APIs are consistent with the Protobuf definitions.
- Use the Makefile for common tasks like building, testing, and running the service.
- Refer to the `README.md` for detailed setup and operational instructions.