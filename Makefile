# vc-sampleapp-go-rust-react — dev tasks
#
# Two ways to run each service:
#   *-local   run on the host (cargo / go / pnpm)
#   *-docker  build the image and run it in a container
#
# Containers reach the DB and each other via the host gateway ($(HOST)), so the
# ports just need to be published on the host. Typical order:
#   make db            # Postgres (detached)
#   make rscounter-*   # local or docker, in its own terminal
#   make goapi-*       # local or docker, in its own terminal
#   make js-*          # local or docker, in its own terminal

# ---- configuration -------------------------------------------------------

# Container runtime. Use `make CONTAINER=docker HOST=host.docker.internal ...`
# to run against Docker instead of Podman.
CONTAINER ?= podman

# Hostname a container uses to reach a port published on the host.
# podman -> host.containers.internal, docker -> host.docker.internal
HOST ?= host.containers.internal

# Image / container names.
RSCOUNTER_IMAGE ?= rscounter
GOAPI_IMAGE     ?= goapi
JS_IMAGE        ?= webapp

# Host ports.
RSCOUNTER_PORT ?= 8081
GOAPI_PORT     ?= 8080
JS_PORT        ?= 3000

# Postgres.
DB_IMAGE     ?= docker.io/library/postgres:18-alpine
DB_CONTAINER ?= rscounter-db
DB_VOLUME    ?= rscounter-pgdata
DB_USER      ?= rscounter
DB_PASSWORD  ?= rscounter
DB_NAME      ?= rscounter
DB_PORT      ?= 5432

.DEFAULT_GOAL := help

# ---- database ------------------------------------------------------------

.PHONY: db db-down db-logs
db: ## Start Postgres (detached)
	$(CONTAINER) run -d --replace --name $(DB_CONTAINER) \
		-e POSTGRES_USER=$(DB_USER) \
		-e POSTGRES_PASSWORD=$(DB_PASSWORD) \
		-e POSTGRES_DB=$(DB_NAME) \
		-p $(DB_PORT):5432 \
		-v $(DB_VOLUME):/var/lib/postgresql \
		--health-cmd "pg_isready -U $(DB_USER)" \
		--health-interval 5s \
		$(DB_IMAGE)

db-down: ## Stop Postgres and remove its volume
	-$(CONTAINER) rm -f $(DB_CONTAINER)
	-$(CONTAINER) volume rm $(DB_VOLUME)

db-logs: ## Follow Postgres logs
	$(CONTAINER) logs -f $(DB_CONTAINER)

# ---- rscounter (Rust) ----------------------------------------------------

.PHONY: rscounter-local rscounter-build rscounter-docker
rscounter-local: ## Run rscounter on the host (needs: make db)
	cd services/rscounter && cargo run

rscounter-build: ## Build the rscounter image
	$(CONTAINER) build -t $(RSCOUNTER_IMAGE) services/rscounter

rscounter-docker: rscounter-build ## Run rscounter in a container (needs: make db)
	$(CONTAINER) run --rm --replace --name $(RSCOUNTER_IMAGE) \
		-p $(RSCOUNTER_PORT):8081 \
		-e DATABASE_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(HOST):$(DB_PORT)/$(DB_NAME) \
		$(RSCOUNTER_IMAGE)

# ---- goapi (Go) ----------------------------------------------------------

.PHONY: goapi-local goapi-build goapi-docker
goapi-local: ## Run goapi on the host (needs rscounter running)
	cd services/goapi && go run .

goapi-build: ## Build the goapi image
	$(CONTAINER) build -t $(GOAPI_IMAGE) services/goapi

goapi-docker: goapi-build ## Run goapi in a container (needs rscounter running)
	$(CONTAINER) run --rm --replace --name $(GOAPI_IMAGE) \
		-p $(GOAPI_PORT):8080 \
		-e RSCOUNTER_URL=http://$(HOST):$(RSCOUNTER_PORT) \
		$(GOAPI_IMAGE)

# ---- js (React) ----------------------------------------------------------

.PHONY: js-install js-local js-build js-docker
js-install: ## Install frontend dependencies
	cd js && pnpm install

js-local: js-install ## Run the React dev server on the host (needs goapi running)
	cd js && pnpm dev

js-build: ## Build the React app image
	$(CONTAINER) build -t $(JS_IMAGE) js

js-docker: js-build ## Run the React app in a container (needs goapi running)
	$(CONTAINER) run --rm --replace --name $(JS_IMAGE) \
		-p $(JS_PORT):3000 \
		-e GOAPI_URL=http://$(HOST):$(GOAPI_PORT) \
		$(JS_IMAGE)

# ---- full stack ----------------------------------------------------------
# The individual *-docker targets run one service at a time; a lone webapp has
# nothing to proxy /api to. `make up` builds and runs the whole stack, wired
# together and in the right order.

.PHONY: db-wait up down
db-wait: ## Block until Postgres reports healthy
	@echo "waiting for postgres to be healthy..."
	@for i in $$(seq 1 30); do \
		if [ "$$($(CONTAINER) inspect -f '{{.State.Health.Status}}' $(DB_CONTAINER) 2>/dev/null)" = "healthy" ]; then \
			echo "postgres is healthy"; exit 0; \
		fi; \
		sleep 1; \
	done; \
	echo "postgres did not become healthy in time" >&2; exit 1

up: db db-wait rscounter-build goapi-build js-build ## Build and run the whole stack (detached)
	$(CONTAINER) run -d --replace --name $(RSCOUNTER_IMAGE) \
		-p $(RSCOUNTER_PORT):8081 \
		-e DATABASE_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(HOST):$(DB_PORT)/$(DB_NAME) \
		$(RSCOUNTER_IMAGE)
	$(CONTAINER) run -d --replace --name $(GOAPI_IMAGE) \
		-p $(GOAPI_PORT):8080 \
		-e RSCOUNTER_URL=http://$(HOST):$(RSCOUNTER_PORT) \
		$(GOAPI_IMAGE)
	$(CONTAINER) run -d --replace --name $(JS_IMAGE) \
		-p $(JS_PORT):3000 \
		-e GOAPI_URL=http://$(HOST):$(GOAPI_PORT) \
		$(JS_IMAGE)
	@echo "stack is up -> http://localhost:$(JS_PORT)"

down: ## Stop and remove all stack containers (keeps the DB volume)
	-$(CONTAINER) rm -f $(JS_IMAGE) $(GOAPI_IMAGE) $(RSCOUNTER_IMAGE) $(DB_CONTAINER)

# ---- housekeeping --------------------------------------------------------

.PHONY: clean help
clean: down ## Stop the stack and force-remove the built images
	-$(CONTAINER) rmi -f $(RSCOUNTER_IMAGE) $(GOAPI_IMAGE) $(JS_IMAGE)

help: ## Show this help
	@grep -E '^[a-zA-Z0-9_-]+:.*?## ' $(MAKEFILE_LIST) \
		| awk 'BEGIN{FS=":.*?## "}{printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'
