# AMS MVP Release Checklist

Use this checklist for an internal MVP release verification from a fresh clone.

## 1. Fresh Clone Setup

- Clone the repository and enter it.
- Confirm local tooling:
  - `go version` uses Go 1.26.2 or a compatible newer Go version.
  - `node --version` uses Node.js 24 or a compatible current LTS.
  - `docker compose version` is available.
- Install frontend dependencies:
  - `cd web && npm ci`

## 2. Environment Setup

- Create local env:
  - `cp .env.example .env`
- Review `.env` values:
  - `DATABASE_URL=postgres://ams:ams@localhost:5432/ams?sslmode=disable`
  - `MINIO_ENDPOINT=localhost:9000`
  - `MINIO_BUCKET=ams-documents`
  - `MAX_UPLOAD_BYTES=20971520`
  - `VITE_DEV_PROXY_TARGET=http://localhost:8080`

## 3. Compose Startup

- Start infrastructure:
  - `docker compose up -d postgres minio`
- Or start the full compose stack:
  - `make up`
- Confirm services:
  - PostgreSQL: `localhost:5432`
  - MinIO API: `localhost:9000`
  - MinIO console: `http://localhost:9001`

## 4. Migrations

- Apply migrations:
  - `make migrate-up`
- For a fresh DB verification, remove compose volumes first:
  - `make down`
  - `docker volume rm ams-ai_postgres-data ams-ai_minio-data`
  - `docker compose up -d postgres minio`
  - `make migrate-up`

## 5. Seed

- Seed demo data:
  - `make seed`
- Confirm demo credentials:
  - Admin: `admin@example.com` / `admin123`
  - Standard user: `user@example.com` / `user123`

## 6. Backend Health Check

- Start backend on the host:
  - `make backend`
- Verify health in another terminal:
  - `curl -fsS http://localhost:8080/healthz`
- Expected response:
  - `{"status":"ok"}`

## 7. Frontend Load

- Start frontend:
  - `make frontend`
- Open:
  - `http://localhost:5173`
- Confirm the login page loads and the Vite proxy reaches the backend.

## 8. Demo Login

- Log in as the admin demo user.
- Confirm the dashboard loads with:
  - total assets
  - upcoming reminders
  - assets by category
  - recently added assets

## 9. Asset Create/Edit Flow

- Create a general asset with:
  - category
  - name
  - brand/model/serial
  - location
  - warranty start and expiry dates
- Open the created asset detail page.
- Edit the asset and confirm the updated values are visible.

## 10. Asset Search/Filter Flow

- Search for the created asset by name or serial number.
- Filter by category.
- Filter by warranty state.
- Filter by document availability after uploading a document.

## 11. Vehicle Create/Edit Flow

- Create a vehicle asset.
- Open the detail page.
- Save a vehicle profile with:
  - registration number
  - vehicle type
  - chassis number
  - engine number
  - odometer
  - next service date or mileage
- Confirm the profile reloads with saved values.

## 12. Service, Fuel, And Renewal Flow

- Add a service record to an asset.
- For a vehicle asset, add:
  - insurance record
  - license record
  - emission/inspection record
  - fuel log
- Confirm records appear in their tables.
- Confirm reminder rows appear when due dates are inside the reminder window.

## 13. Document Upload/Download/Delete

- Upload a JPG, PNG, or PDF document smaller than `MAX_UPLOAD_BYTES`.
- Confirm unsupported files are rejected with a clear error.
- Download the document through the app.
- Confirm unauthenticated direct download requests return `401`.
- Delete the document and confirm it disappears from the asset detail page.

## 14. Reminder Visibility

- Open `/reminders`.
- Confirm warranty, vehicle renewal, and service reminders are visible for seeded or newly created records inside the configured reminder window.

## 15. Reports / Export

- Open `/reports`.
- Download each implemented CSV export:
  - asset list
  - warranty expiry report
  - vehicle renewal report
  - service history report
  - fuel log export
- Confirm each file downloads and contains headers.

## 16. Test, Lint, Format, And Contract Verification

- Run:
  - `make fmt`
  - `make test`
  - `make lint`
  - `make openapi-validate`
- Install the Playwright browser once per machine or CI image:
  - `make e2e-install`
- Run the MVP browser smoke test:
  - `make e2e`

## 17. Backup / Restore Notes

- Local PostgreSQL backup:
  - `docker compose exec postgres pg_dump -U ams -d ams > ams_backup.sql`
- Local PostgreSQL restore into an empty database:
  - `cat ams_backup.sql | docker compose exec -T postgres psql -U ams -d ams`
- Local MinIO data is stored in the `minio-data` compose volume. For local MVP verification, preserve or snapshot that volume together with the PostgreSQL backup if uploaded documents must be retained.

## 18. Known Limitations / Deferred Items

- AI, OCR, chat search, and predictive maintenance are intentionally out of MVP scope.
- Reminders are in-app only; email notifications are deferred.
- CSV export is implemented; PDF export and CSV import are deferred.
- Assignment history is deferred; only current assignment fields are stored.
- Document storage is private through MinIO and backend-proxied downloads; external object lifecycle policies are not configured for local development.
- There is no production deployment automation in this MVP release checklist.
