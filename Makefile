build:
	go build ./cmd/firmdiff

test:
	go test ./...

lint:
	golangci-lint run

run:
	go run ./cmd/firmdiff