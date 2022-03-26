lint:
	@staticcheck ./...

build:
	@cd cmd/api && go build

test:
	@go test -v ./... -tags test

check: lint
