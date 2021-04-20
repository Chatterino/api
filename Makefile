lint:
	@staticcheck ./...

build:
	@cd cmd/api && go build

check: lint
