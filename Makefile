.PHONY: backend frontend test docker-up docker-down docker-build docker-logs

backend:
	cd backend && go run ./cmd/server

frontend:
	cd frontend && npm run dev

test:
	cd backend && go test ./... -cover
	cd frontend && npm test -- --run

docker-up:
	docker compose up --build

docker-down:
	docker compose down

docker-build:
	docker compose build

docker-logs:
	docker compose logs -f