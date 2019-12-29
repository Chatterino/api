lint:
	@golangci-lint run

lint-current:
	@golangci-lint run --new

build:
	@go build

check: lint
