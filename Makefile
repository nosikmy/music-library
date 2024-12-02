ifneq (,$(wildcard ./.env))
	include .env
	export
endif

.PHONY: migrationUp
migrationUp:
	docker run --rm -v D:\Ucheba\music-library\migrations:/migrations --network appnet migrate/migrate -path=./migrations \
    		-database 'postgres://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)' up

.PHONY: migrationDown
migrationDown:
	docker run --rm -v D:\Ucheba\music-library\migrations:/migrations --network appnet migrate/migrate -path=./migrations \
    		-database 'postgres://$(DB_USERNAME):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)' down


.PHONY: up
up:
	docker compose up -d --build

.PHONY: down
down:
	docker compose down -v