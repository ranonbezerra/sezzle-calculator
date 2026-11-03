.PHONY: test backend-test backend-lint backend-run backend-cover frontend-test frontend-lint frontend-run frontend-cover

test: backend-test frontend-test

backend-test:
	cd backend && go test -race ./...

backend-lint:
	cd backend && go vet ./... && golangci-lint run ./...

backend-run:
	cd backend && go run ./cmd/server

backend-cover:
	cd backend && go test -race -coverprofile=coverage.out ./... && go tool cover -func=coverage.out

frontend-test:
	cd frontend && npm run coverage

frontend-lint:
	cd frontend && npm run lint && npm run typecheck

frontend-run:
	cd frontend && npm run dev

frontend-cover:
	cd frontend && npm run coverage
