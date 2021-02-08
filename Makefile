include .env

DEFAULT_GOAL := help
help:
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z0-9_-]+:.*?##/ { printf "  \033[36m%-27s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

build: ## Build compressed project.
	GOOS=linux go build -o bin/$(PROJECT_NAME) main.go

bc: ## Build and copy to server.
	go build -ldflags "-s -w" -o bin/$(PROJECT_NAME)
	scp bin/$(PROJECT_NAME) $(SERVER_LOGIN):$(SERVER_DIR)

build-image: ## Build docker image
	docker build -t mamau/restream:latest .
	docker push mamau/restream

up: ## Start project
	docker-compose up -d

up-fresh: ## Start project
	docker-compose down
	make build-image
	docker-compose up