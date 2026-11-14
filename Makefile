.PHONY: backend frontend test test-backend test-frontend lint coverage docker-up docker-down docker-build docker-logs

test-backend:
	cd backend && go test ./... -race -cover

test-frontend:
	cd frontend && npm run test:run

lint:
	cd backend && go vet ./...
	cd frontend && npm run lint

coverage:
	cd backend && go test -race ./... -coverprofile coverage.out
	cd backend && go tool cover -html=coverage.out -o coverage.html
	cd frontend && npm run test:coverage

backend:
	cd backend && go run ./cmd/server

frontend:
	cd frontend && npm run dev

test:
	$(MAKE) test-backend
	$(MAKE) test-frontend

docker-up:
	docker compose up --build

docker-down:
	docker compose down

docker-build:
	docker compose build

docker-logs:
	docker compose logs -f