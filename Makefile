SHELL := /bin/sh

ifneq (,$(wildcard .env))
include .env
export
endif

.DEFAULT_GOAL := help

.PHONY: help up down logs dev backend frontend test lint fmt openapi-validate e2e-install e2e migrate-up migrate-down seed

help:
	@printf "%s\n" "AMS development targets:"
	@printf "%s\n" "  make up            Build and start postgres, minio, backend, and frontend"
	@printf "%s\n" "  make down          Stop the local stack"
	@printf "%s\n" "  make logs          Follow docker compose logs"
	@printf "%s\n" "  make dev           Start infra, run migrations, seed data, and show URLs"
	@printf "%s\n" "  make backend       Run backend on the host"
	@printf "%s\n" "  make frontend      Run frontend on the host"
	@printf "%s\n" "  make test          Run backend and frontend tests"
	@printf "%s\n" "  make lint          Run frontend lint and backend vet"
	@printf "%s\n" "  make fmt           Format Go and frontend code"
	@printf "%s\n" "  make openapi-validate  Check OpenAPI YAML and backend route coverage"
	@printf "%s\n" "  make e2e-install   Install the Playwright Chromium browser"
	@printf "%s\n" "  make e2e           Run the MVP browser smoke test"
	@printf "%s\n" "  make migrate-up    Apply database migrations"
	@printf "%s\n" "  make migrate-down  Roll back latest migration"
	@printf "%s\n" "  make seed          Seed demo data"

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f

dev: up migrate-up seed
	@printf "%s\n" "Frontend: http://localhost:5173"
	@printf "%s\n" "Backend:  http://localhost:8080/healthz"
	@printf "%s\n" "MinIO:    http://localhost:9001"

backend:
	go run ./cmd/api

frontend:
	cd web && npm run dev

test:
	go test ./cmd/... ./internal/...
	cd web && npm test

lint:
	go vet ./cmd/... ./internal/...
	cd web && npm run lint

fmt:
	gofmt -w cmd internal
	cd web && npm run fmt

openapi-validate:
	go test ./internal/httpapi -run TestOpenAPIContractMatchesRouter

e2e-install:
	cd web && npm run e2e:install

e2e:
	docker compose up -d postgres minio
	docker compose stop backend frontend
	go run ./cmd/migrate up
	go run ./cmd/seed
	@set -e; \
		go run ./cmd/api > /tmp/ams-api-e2e.log 2>&1 & \
		api_pid=$$!; \
		cleanup() { kill $$api_pid >/dev/null 2>&1 || true; wait $$api_pid 2>/dev/null || true; }; \
		trap cleanup EXIT INT TERM; \
		ready=0; \
		for _ in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15; do \
			if curl -fsS http://localhost:8080/healthz >/dev/null 2>&1; then ready=1; break; fi; \
			sleep 1; \
		done; \
		if [ "$$ready" != "1" ]; then cat /tmp/ams-api-e2e.log; exit 1; fi; \
		cd web && npm run e2e

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

seed:
	go run ./cmd/seed
