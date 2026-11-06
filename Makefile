.PHONY: all build build-frontend icons test test-race clean dev

BINARY_NAME=applauncher3
OUTPUT_DIR=bin

all: build-frontend build

icons:
	python3 scripts/generate_icons.py

build-frontend:
	cd frontend && npm run build

build:
	mkdir -p $(OUTPUT_DIR)
	go build -tags "production,webkit2_41" -o $(OUTPUT_DIR)/$(BINARY_NAME) main.go app.go

build-windows: build-frontend
	mkdir -p $(OUTPUT_DIR)
	GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -tags "production" -ldflags "-H windowsgui" -o $(OUTPUT_DIR)/$(BINARY_NAME).exe main.go app.go

test:
	go test -v ./...
	go test -v tests/e2e/e2e_test.go
	cd frontend && npm test

test-race:
	go test -v -race -timeout 30s ./...
	go test -v -race tests/e2e/e2e_test.go
	cd frontend && npm test

clean:
	rm -rf $(OUTPUT_DIR) frontend/dist
