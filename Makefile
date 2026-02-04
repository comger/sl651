.PHONY: all build run clean test import

all: build

build:
	go build -o bin/sl651-platform cmd/server/main.go
	go build -o bin/importer cmd/importer/main.go

run:
	go run cmd/server/main.go

import:
	go run cmd/importer/main.go testdata.csv

test:
	go test -v ./...

test-full:
	@echo "运行完整测试..."
	@./test.sh

clean:
	rm -rf bin/
	rm -rf data/
	rm -rf logs/

deps:
	go mod download
	go mod tidy

fmt:
	go fmt ./...

vet:
	go vet ./...
