DB_URL=postgres://dev:devpass@localhost:5432/purchases?sslmode=disable

up:
	docker-compose up -d

down:
	docker-compose down

migrate-up:
	migrate -path ./internal/db/migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path ./internal/db/migrations -database "$(DB_URL)" down

run:
	go run ./cmd/main.go