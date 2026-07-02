## InternlyApp
---
Stephen Sulimani, 2026

### API tests

```bash
go test ./...
```

Route tests that hit Postgres require `TEST_DATABASE_URL`. Without it, those cases are skipped and unit tests still run. Each database test runs inside a transaction that is rolled back, so no rows persist.

```bash
docker compose up -d db
export TEST_DATABASE_URL='postgres://USER:PASS@localhost:5432/internly?sslmode=disable'
go test ./cmd/api/routes/... -v
```
