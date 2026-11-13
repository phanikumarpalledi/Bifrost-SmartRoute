.PHONY: all build build-frontend build-backend run test bench demo docker clean

PORT ?= 8080

all: build

build-frontend:
	@echo "Building React dashboard..."
	cd frontend && npm install && npm run build

build-backend:
	@echo "Building Go gateway binary..."
	cd backend && go build -o ../bifrost-smartroute ./cmd/server

build: build-frontend build-backend
	@echo "Build complete."

run:
	@echo "Starting Bifrost SmartRoute on port $(PORT)..."
	./bifrost-smartroute -port $(PORT)

test:
	@echo "Running all backend tests..."
	cd backend && go test -v ./...

bench:
	@echo "Running concurrent benchmark tool..."
	go run scripts/benchmark.go -total 50 -concurrency 10

demo:
	@echo "Running interactive CLI demo..."
	python scripts/demo.py

docker:
	@echo "Building and launching Docker containers..."
	cd docker && docker-compose up --build -d

clean:
	@echo "Cleaning up build artifacts..."
	rm -f bifrost-smartroute bifrost-smartroute.exe
	rm -f *.db *.db-journal
