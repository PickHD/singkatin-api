build:
	docker compose build
run:
	docker compose up -d --build --force-recreate
stop:
	docker compose down
remove:
	docker compose down -v
	
.PHONY: build run stop remove

