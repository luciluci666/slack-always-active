# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o slack-always-active

# Final stage
FROM alpine:3.19

# Add non root user
RUN adduser -D -g '' appuser

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /app

# Create logs directory with proper permissions
RUN mkdir -p /app/logs && chown -R appuser:appuser /app/logs
RUN mkdir -p /app/cache/cache && chown -R appuser:appuser /app/cache/cache

# Copy binary from builder
COPY --from=builder /app/slack-always-active .
COPY --from=builder /app/.env.example .env

# Use non root user
USER appuser

# Create volume for logs
VOLUME ["/app/logs"]

# Run the application
CMD ["./slack-always-active"] 