# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Development Commands

### Go Backend

- `make build` - Build the main server binary
- `make lint` - Run goimports and golangci-lint
- `make unit-tests` - Run unit tests (excludes e2e)
- `go test -v -count=1 -race ./e2e/...` - Run e2e tests directly

### Frontend (SolidJS)

- `cd web && npm run local` - Run frontend with local backend
- `cd web && npm run dev` - Run frontend only (requires proxy config)
- `cd web && npm run build` - Build production frontend

### Environment Management

- `make start-e2e-env` - Start complete local dev environment with Docker
- `make run-e2e-tests` - Run full e2e test suite
- `make start-e2eci-env` - Start CI-friendly e2e environment
- `make run-staging-env` - Start staging environment

### Database

- `make migrate_up` - Apply database migrations
- `make migrate_down` - Rollback migrations
- `make migrate_new MNAME=migration_name` - Create new migration

## Architecture Overview

**Treenq** is an open-source Platform-as-a-Service for Kubernetes with a clean architecture pattern:

### Core Structure

```
src/
├── api/           # Application bootstrapping and HTTP configuration
├── domain/        # Business logic handlers (Deploy, Auth, Secrets, GitHub integration)
├── repo/          # Data access layer (Store, Git, GitHub, Artifacts, Extract)
├── resources/     # Infrastructure components (DB, router setup)
└── services/      # External integrations (Auth, CDK/Kubernetes)

pkg/
├── crypto/        # Signature verification utilities
├── sdk/           # Treenq SDK types and space definitions
└── vel/           # Custom web framework (routing, auth, logging)
```

### Key Components

**Domain Handlers**: Core business logic in `src/domain/` implementing clean architecture principles. The `Handler` struct coordinates all operations with dependency injection through interfaces.

**Repository Layer**: `src/repo/` provides data access abstractions:

- `Store`: PostgreSQL operations with Squirrel query builder
- `Git`: Repository cloning and management
- `GitHub`: GitHub API integration and app installations
- `Artifacts`: Container image building (uses BuildKit)
- `Extract`: Configuration extraction from repositories

**Database Schema**: PostgreSQL with migrations tracking users, GitHub installations, repositories, deployments (with JSONB space config), and secrets.

**Container Building**: Recently migrated from Buildah to BuildKit for improved performance and caching.

## Development Setup

**Local Development**: Use `docker compose up` which starts PostgreSQL, Docker registry, BuildKit, and K3s in containers.

**Frontend Development**: Two modes available - full local stack or proxy-only development targeting staging environment.

## Testing Strategy

**Unit Tests**: Package-level testing focusing on business logic
**E2E Tests**: Full integration testing with real Kubernetes cluster and database
**Test Data**: Embedded fixtures and webhook payloads in test files

## Configuration

Environment-based configuration using `envconfig` for:

- GitHub integration (OAuth, webhooks, private keys)
- Docker registry authentication
- Database connections
- Kubernetes cluster access
- BuildKit host configuration
- JWT token management

the project uses multiple docker compose configurations for different environments (dev, e2e, staging).
