.PHONY: build run swagger tidy

BINARY=vectovm-api

build:
	go build -o bin/$(BINARY) ./cmd/vectovm-api

run: build
	./bin/$(BINARY)

swagger:
	GOPATH=/tmp/vectovm-gopath GOROOT=$$(go env GOROOT) GOFLAGS=-mod=mod go run github.com/swaggo/swag/cmd/swag@v1.16.4 init -g main.go -o api/docs -d cmd/vectovm-api,internal/server/http/handler,internal/model --parseInternal --parseDependency

tidy:
	go mod tidy
