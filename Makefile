.PHONY: backend frontend test

backend:
	cd backend && go run ./cmd/server

frontend:
	cd frontend && npm run dev

test:
	cd backend && go test ./... -cover
	cd frontend && npm test -- --run