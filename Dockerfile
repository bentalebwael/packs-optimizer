# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o server ./cmd/server/main.go

# Runtime stage
FROM alpine:3.16

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/server ./server

# Copy static files and migrations
COPY static/ ./static/
COPY migrations/ ./migrations/

# Create non-root user
RUN adduser -D appuser
USER appuser

# Define environment variable with a default value (optional)
ENV DATABASE_URL=""

# Expose port
EXPOSE 8080

# Command to run the server
CMD ["./server"]