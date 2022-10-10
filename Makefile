lint:
	@staticcheck ./...

build:
	@cd cmd/api && go build

test:
	@go test ./... -tags test

cover:
	@go test ./... -cover -tags test

cover_html:
	@go test ./... -coverprofile=coverage.out -tags test && go tool cover -html=coverage.out

vtest:
	@go test -v ./... -tags test

check: lint
