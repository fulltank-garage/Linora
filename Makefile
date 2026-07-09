.PHONY: install dev-web dev-api build test-api db-up db-down

install:
	pnpm install

dev-web:
	pnpm dev:web

dev-api:
	pnpm dev:api

build:
	pnpm build

test-api:
	cd apps/api && go test ./...

db-up:
	pnpm db:up

db-down:
	pnpm db:down
