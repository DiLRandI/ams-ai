SHELL := /bin/sh

ifneq (,$(wildcard .env))
include .env
export
endif

.DEFAULT_GOAL := help

.PHONY: help up down logs dev backend frontend test lint fmt migrate-up migrate-down seed

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

migrate-up:
	go run ./cmd/migrate up

migrate-down:
	go run ./cmd/migrate down

seed:
	go run ./cmd/seed
