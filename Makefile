.PHONY: migrate-up test-auth test-product test

migrate-up:
	docker compose -f services/docker-compose.yml up -d postgres
	docker compose -f services/docker-compose.yml up migrate-auth migrate-product

test-auth:
	cd services/auth-service && go test ./internal/config ./internal/delivery/... ./internal/usecase -coverprofile=coverage.out

test-product:
	cd services/product-service && go test ./internal/config ./internal/delivery/... ./internal/usecase -coverprofile=coverage.out

test: test-auth test-product
