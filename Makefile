.PHONY: build up down logs migrate ps

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