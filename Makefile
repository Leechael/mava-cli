.PHONY: build test lint clean install

BINARY_NAME := mava
BUILD_TARGET := ./cmd/mava
BIN_DIR := bin

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(BINARY_NAME) $(BUILD_TARGET)

test:
	go test ./...

lint:
	go vet ./...

clean:
	rm -rf $(BIN_DIR)

install:
	go install $(BUILD_TARGET)
