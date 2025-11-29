.PHONY: build up down logs migrate ps tidy restart swagger stop-app db

build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down

logs:
	docker compose logs -f app

migrate:
	docker compose run --rm migrator

ps:
	docker compose ps

tidy:
	go mod tidy

swagger:
	swag init -g cmd/tiktok/main.go -o internal/api/docs

stop-app:
	docker compose stop app migrator

db:
	docker compose up -d postgres
restart:
	$(MAKE) swagger
	$(MAKE) down
	$(MAKE) build
	$(MAKE) up
	$(MAKE) logs
