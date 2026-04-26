COMPOSE_BASE = docker compose -f docker-compose.yml
COMPOSE_LOCAL = docker compose -f docker-compose.yml -f docker-compose.local.yml
ENV_FILE ?= .env

up:
	$(COMPOSE_BASE) --env-file $(ENV_FILE) up -d

up-local:
	$(COMPOSE_LOCAL) --env-file $(ENV_FILE) up -d

down:
	$(COMPOSE_BASE) --env-file $(ENV_FILE) down

down-local:
	$(COMPOSE_LOCAL) --env-file $(ENV_FILE) down

restart: down up

pull:
	$(COMPOSE_BASE) --env-file $(ENV_FILE) pull

logs:
	$(COMPOSE_BASE) --env-file $(ENV_FILE) logs -f --tail=200

ps:
	$(COMPOSE_BASE) --env-file $(ENV_FILE) ps

migrate:
	$(COMPOSE_BASE) --env-file $(ENV_FILE) run --rm migrate

migrate-local:
	$(COMPOSE_LOCAL) --env-file $(ENV_FILE) run --rm migrate

generate:
	@make -C ./backend generate
	@make -C ./frontend generate

