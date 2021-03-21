lint:
	@staticcheck ./...

build:
	@go build

check: lint
