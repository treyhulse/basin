# Build stage
FROM golang:1.21-alpine AS builder

# Set build arguments
ARG BUILD_ENV=production
ARG GOOS=linux
ARG GOARCH=amd64

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application with environment-specific flags
RUN CGO_ENABLED=0 GOOS=${GOOS} GOARCH=${GOARCH} go build \
    -a -installsuffix cgo \
    -ldflags="-w -s" \
    -o main cmd/main.go

# Production stage
FROM alpine:latest AS production

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy migrations directory
COPY --from=builder /app/migrations ./migrations

# Copy any other necessary files
COPY --from=builder /app/public ./public

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Change ownership
RUN chown -R appuser:appgroup /app

USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the binary
CMD ["./main"]

# Development stage
FROM golang:1.21-alpine AS development

# Install development dependencies
RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application for development (with debug info)
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -o main cmd/main.go

# Copy migrations directory
COPY --from=builder /app/migrations ./migrations

# Copy any other necessary files
COPY --from=builder /app/public ./public

# Expose port
EXPOSE 8080

# Run the binary
CMD ["./main"]
