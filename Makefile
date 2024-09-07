.DEFAULT_GOAL := help

GOMARKDOC := $(shell command -v gomarkdoc 2> /dev/null)

## Run linter. Usage: 'make lint'
lint: ; $(info Running lint...)
	@golangci-lint run

## Run linter. Usage: 'make lint-fix'
lint-fix: ; $(info Running lint fix...)
	@golangci-lint run --fix

## Run tests. Usage: 'make test'
test: ; $(info running tests…)
	@go test -race -failfast -timeout 3m

## Generate the docs. Usage: 'make docs'
docs: ; $(info Generating docs...)
ifndef GOMARKDOC
	@go install github.com/princjef/gomarkdoc/cmd/gomarkdoc@v1.1.0
endif
	@gomarkdoc ./...

# -- help

# COLORS
GREEN  := $(shell tput -Txterm setaf 2)
YELLOW := $(shell tput -Txterm setaf 3)
WHITE  := $(shell tput -Txterm setaf 7)
RESET  := $(shell tput -Txterm sgr0)
TARGET_MAX_CHAR_NUM=20
## Show help
help:
	@echo ''
	@echo 'Usage:'
	@echo '  ${YELLOW}make${RESET} ${GREEN}<target>${RESET}'
	@echo ''
	@echo 'Targets:'
	@awk '/^[a-zA-Z\-\_0-9]+:/ { \
		helpMessage = match(lastLine, /^## (.*)/); \
		if (helpMessage) { \
			helpCommand = substr($$1, 0, index($$1, ":")); \
			helpMessage = substr(lastLine, RSTART + 3, RLENGTH); \
			printf "  ${YELLOW}%-$(TARGET_MAX_CHAR_NUM)s${RESET} ${GREEN}%s${RESET}\n", helpCommand, helpMessage; \
		} \
	} \
	{ lastLine = $$0 }' $(MAKEFILE_LIST)
