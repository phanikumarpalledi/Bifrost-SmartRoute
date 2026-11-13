# Stage 1: Build the React Dashboard UI
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# Stage 2: Build the Go Bifrost Gateway
FROM golang:1.23-alpine AS backend-builder
WORKDIR /app
RUN apk add --no-cache git ca-certificates tzdata

# Copy Bifrost core dependency and backend source
COPY third_party/ ./third_party/
COPY backend/ ./backend/
WORKDIR /app/backend
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/bifrost-smartroute ./cmd/server

# Stage 3: Minimal Production Image
FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata curl
WORKDIR /app

# Copy executable and compiled static UI
COPY --from=backend-builder /app/bifrost-smartroute /app/bifrost-smartroute
COPY --from=frontend-builder /app/frontend/dist /app/frontend/dist

# Expose Gateway & Dashboard Port
EXPOSE 8080

# Environment Defaults
ENV PORT=8080
ENV DEMO_MODE=true

ENTRYPOINT ["/app/bifrost-smartroute"]
CMD ["-port", "8080", "-db", "/app/data/bifrost_smartroute.db"]
