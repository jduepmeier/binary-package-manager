.PHONY: clean test-coverage test check-env
PREFIX := $(GOPATH)
INSTALL_PATH := $(PREFIX)/bin

build: bin go.mod go.sum *.go cmd/bpm/*.go
	go build -o bin/bpm cmd/bpm/*.go

install: check-env build
	install bin/bpm $(INSTALL_PATH)/bpm

bin:
	mkdir -p bin

clean:
	rm -rf bin

test:
	go test -cover -tags test ./...

test-coverage: bin
	go test -coverprofile=bin/coverage.out -tags test ./...
	go tool cover -html bin/coverage.out

check-env:
ifndef PREFIX
	$(error PREFIX or GOPATH not set. Cannot install into /bin)
endif
