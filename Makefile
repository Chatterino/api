lint:
	@staticcheck ./...

build:
	@cd cmd/api && go build

test:
	@go test ./... -tags test

vtest:
	@go test -v ./... -tags test

check: lint
