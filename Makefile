COMPOSE ?= DOCKER_BUILDKIT=1 COMPOSE_DOCKER_CLI_BUILD=1 docker compose
SHELLCHECK ?= docker run --rm -it -v $(PWD):$(PWD) -w $(PWD) koalaman/shellcheck-alpine shellcheck

all:
	$(COMPOSE) up --force-recreate --build --abort-on-container-exit --remove-orphans

debug:
	LOG_LEVEL=debug $(MAKE) all

pgcli:
	pgcli postgresql://jump:password@localhost:5432/jump?sslmode=disable

lint:
	$(SHELLCHECK) $$(find . -name '*.sh' -not -path "*/_build/*")
	$(COMPOSE) config --quiet

update:
	cd srv && go get -u -a && go mod tidy && go mod verify

test:
	cd srv && go vet ./...
	curl -fsS -X GET -H 'Accept: application/json' localhost:8088/users | grep -F '{"user_id":4,'
	rm -f /tmp/bla && touch /tmp/bla
	curl -fsS -o /tmp/bla -X POST -H 'Accept: application/json' -H 'Content-Type: application/json' -d '{"user_id":4, "amount":113.45, "label":"Work for April"}' localhost:8088/invoice
	curl -fsS -X POST -H 'Accept: application/json' -H 'Content-Type: application/json' -d '{"invoice_id":'$$(cat /tmp/bla | jq .invoice_id)', "amount":113.45, "reference":"JMPINV200220117"}' localhost:8088/transaction
	cd srv && go test ./...

clean:
	$(COMPOSE) down
