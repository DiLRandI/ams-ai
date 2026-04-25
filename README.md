# Asset Management System MVP

AMS is a local-development MVP for tracking physical assets, vehicle assets, documents, warranties, renewals, service history, fuel logs, and in-app reminders.

The MVP follows the project docs in `docs/prd_asset_management_system.md` and `docs/feature_list_asset_management_system.md`. It intentionally excludes AI/OCR, digital assets, subscriptions, procurement, native mobile apps, and enterprise workflows.

## Architecture

- Frontend: React, TypeScript, Vite, React Router, TanStack Query, React Hook Form.
- Backend: Go with the standard library `net/http` `ServeMux`.
- Database: PostgreSQL with explicit SQL migrations.
- File storage: S3-compatible object storage through MinIO for local development.
- API: REST JSON, with private document downloads proxied through the backend.

Main folders:

- `cmd/api`: backend server entrypoint.
- `cmd/migrate`: migration runner.
- `cmd/seed`: demo data seeder.
- `internal`: backend config, HTTP, service, repository, domain, and storage code.
- `migrations`: PostgreSQL schema migrations.
- `web`: React frontend.
- `docs/openapi.yaml`: implemented API contract.

## Local Setup

Requirements:

- Go 1.26.2 or compatible newer Go version
- Node.js 24 or compatible current LTS
- Docker and Docker Compose

Create environment configuration:

```sh
cp .env.example .env
```

Install frontend dependencies:

```sh
cd web && npm ci
```

Start infrastructure, apply migrations, seed demo data, and run the app on the host:

```sh
docker compose up -d postgres minio
make migrate-up
make seed
make backend
```

In another terminal:

```sh
make frontend
```

Or start the full compose stack and then apply the database setup from the host:

```sh
make up
make migrate-up
make seed
```

Local URLs:

- Frontend: http://localhost:5173
- Backend health: http://localhost:8080/healthz
- MinIO console: http://localhost:9001

Demo credentials:

- Admin: `admin@example.com` / `admin123`
- Standard user: `user@example.com` / `user123`

## Common Commands

```sh
make help
make up
make down
make logs
make backend
make frontend
make test
make lint
make fmt
make openapi-validate
make e2e-install
make e2e
make migrate-up
make migrate-down
make seed
```

Typical host-development workflow:

```sh
docker compose up -d postgres minio
make migrate-up
make seed
make backend
```

In another terminal:

```sh
make frontend
```

Fresh local database reset:

```sh
make down
docker volume rm ams-ai_postgres18-data ams-ai_minio-data
docker compose up -d postgres minio
make migrate-up
make seed
```

## MVP Scope Implemented

- Email/password login with `admin` and `user` roles.
- Protected backend API routes and protected frontend routes.
- Asset CRUD, archive, restore, categories, status, location, assignment fields, warranty fields, search, and filters.
- Document upload/list/download/delete for JPG, PNG, and PDF files.
- Vehicle profile with registration, odometer, service due date/mileage, driver, chassis, and engine fields.
- Vehicle insurance, license, and emission/inspection records with cost and expiry dates.
- Service/repair records for all assets.
- Fuel logs for vehicle assets.
- Dashboard summaries and in-app reminders.
- CSV exports for assets, warranties, vehicle renewals, service history, and fuel logs.

## MVP Decisions From Ambiguities

- Reminder window is fixed at 30 days for MVP through `REMINDER_WINDOW_DAYS`.
- Reminders are in-app only. Email reminders are deferred because the docs mark them optional/open.
- CSV export is included. PDF export and CSV import are deferred.
- Assignment history is not included; the MVP stores the current assigned person and optional assigned app user.
- Emission/inspection records are included because the feature list marks them as MVP must-have.
- Standard users can work with assets they created or that are assigned to them. Admins can access all records and manage categories.

## Documents And Storage

Uploaded files are stored in MinIO under deterministic private object keys:

```text
assets/{asset_id}/documents/{timestamp}/{safe_filename}
```

Documents are downloaded through the backend so the MinIO bucket does not need public access.
Direct unauthenticated document downloads return `401`; asset-level access rules are checked before downloads and deletes.

Supported file types:

- JPG/JPEG
- PNG
- PDF

Default max upload size is 20 MB.
The limit is configured with `MAX_UPLOAD_BYTES`. The backend sniffs uploaded file contents and only stores files whose detected MIME type is `image/jpeg`, `image/png`, or `application/pdf`.

## API Contract

The OpenAPI contract lives at `docs/openapi.yaml`.

Validate the OpenAPI YAML and route coverage against the backend router:

```sh
make openapi-validate
```

## Testing

Run all tests:

```sh
make test
```

Backend tests cover core domain rules and key HTTP behavior. Frontend tests cover important UI flows and rendering.

Run lint checks:

```sh
make lint
```

Run the browser smoke test after PostgreSQL and MinIO are available:

```sh
make e2e-install
make e2e
```

`make e2e` applies migrations, seeds demo data, starts the backend temporarily, starts the Vite dev server through Playwright, and verifies the core MVP happy path.

## Release Checklist

Use `docs/RELEASE_CHECKLIST.md` for the internal MVP release verification pass, including fresh clone setup, migrations, seed, smoke flows, document authorization, reports, backup notes, and known limitations.
