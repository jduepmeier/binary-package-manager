.PHONY: clean test-coverage test

build: bin go.mod go.sum *.go cmd/bpm/*.go
	go build -o bin/bpm cmd/bpm/*.go

bin:
	mkdir -p bin

clean:
	rm -rf bin

test:
	go test -cover

test-coverage: bin
	go test -coverprofile=bin/coverage.out
	go tool cover -html bin/coverage.out