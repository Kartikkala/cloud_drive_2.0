# Makefile
.PHONY: build run clean dev

build:
	@echo "Building binary..."
	@go build -o bin/cloud-drive cmd/server/main.go

run: build
	@echo "Running..."
	@./bin/cloud-drive

clean:
	@rm -rf bin/

dev:
	@echo "Starting dev server..."
	@/home/sirkartik/go/bin/air
