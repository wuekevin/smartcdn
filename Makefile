.PHONY: build frontend test clean dev

# Build the production binary with embedded frontend
build: frontend
	go build -ldflags "-s -w" -o dist/smartcdn ./cmd/smartcdn/

# Build only the frontend
frontend:
	cd frontend && npm run build
	rm -rf assets/dist
	cp -r frontend/dist assets/dist

# Run all tests
test:
	go test ./...

# Run tests with race detector
test-race:
	go test -race ./...

# Run the dev server (requires SMARTCDN_ENCRYPT_KEY)
dev:
	SMARTCDN_ENCRYPT_KEY=$${SMARTCDN_ENCRYPT_KEY:-0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20} \
	go run ./cmd/smartcdn/

clean:
	rm -rf dist/ assets/dist frontend/dist
