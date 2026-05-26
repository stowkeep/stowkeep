run:
	@go run ./cmd/stowkeep

dev:
	@mkdir -p .data
	@$(MAKE) -j2 dev-api dev-web

dev-api:
	@STOWKEEP_LOG_LEVEL=debug STOWKEEP_LOG_FORMAT=text \
	 STOWKEEP_DATABASE_DRIVER=sqlite STOWKEEP_DATABASE_PATH=./.data/dev.db \
	 go run ./cmd/stowkeep

dev-web:
	@cd web && npm run dev

dev-postgres:
	@docker compose -f docker-compose.dev.yml up -d postgres
	@STOWKEEP_LOG_LEVEL=debug STOWKEEP_LOG_FORMAT=text \
	 STOWKEEP_DATABASE_DRIVER=postgres \
	 STOWKEEP_DATABASE_URL=postgres://stowkeep:stowkeep@localhost:5432/stowkeep?sslmode=disable \
	 go run ./cmd/stowkeep

build: build-web
	@CGO_ENABLED=0 go build -ldflags "-s -w" -o bin/stowkeep ./cmd/stowkeep

build-web:
	@cd web && npm run build

test: test-go test-web

test-go:
	@go test ./cmd/... ./pkg/... -short -race -count=1

test-web:
	@cd web && npm run test

test-integration:
	@go test ./... -race -count=1 -tags=integration

test-migrations-sqlite:
	@go test ./pkg/db/... -run TestMigrationsSQLite -count=1

test-migrations-postgres:
	@STOWKEEP_DATABASE_URL=$${STOWKEEP_DATABASE_URL:-postgres://stowkeep:stowkeep@localhost:5432/stowkeep?sslmode=disable} \
	 go test ./pkg/db/... -run TestMigrationsPostgresFromEnv -count=1

lint: lint-go lint-web

lint-go:
	@golangci-lint run ./...

lint-web:
	@cd web && npm run lint && npm run typecheck

migrate-up:
	@go run ./cmd/stowkeep --migrate-only 2>/dev/null || goose -dir migrations/sqlite sqlite "$${STOWKEEP_DATABASE_PATH:-./.data/dev.db}" up

docker-build:
	@docker build -t stowkeep:local .

.PHONY: run dev dev-api dev-web dev-postgres build build-web test test-go test-web test-integration \
	test-migrations-sqlite test-migrations-postgres lint lint-go lint-web migrate-up docker-build
