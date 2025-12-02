.PHONY: build up down logs migrate ps tidy restart swagger stop-app db run

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


run:
	$(MAKE) swagger
	$(MAKE) build
	$(MAKE) up
	$(MAKE) migrate
	$(MAKE) logs

restart:
	$(MAKE) down
	$(MAKE) run
