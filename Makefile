.PHONY: install backend frontend test test-backend test-frontend lint coverage build docker-up docker-down docker-build docker-logs

install:
	cd frontend && npm ci

test-backend:
	cd backend && go test ./... -race -cover

test-frontend:
	cd frontend && npm run test:run

lint:
	cd backend && go vet ./...
	cd frontend && npm run lint

coverage:
	cd backend && go test ./... -coverprofile coverage.out
	cd backend && go tool cover -html=coverage.out -o coverage.html
	cd frontend && npm run test:coverage

build:
	cd frontend && npm run build
	cd backend && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" ./cmd/server

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