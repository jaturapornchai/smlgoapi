# Build stage
FROM golang:1.24-alpine AS builder

# Install git and ca-certificates
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o smlgoapi .

# Final stage
FROM alpine:latest

# Install ca-certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata

# Set timezone to Asia/Bangkok
ENV TZ=Asia/Bangkok

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/smlgoapi .

# Copy necessary directories
COPY --from=builder /app/fonts ./fonts
COPY --from=builder /app/config ./config
COPY --from=builder /app/sql ./sql

# Create image_cache directory and set ownership
RUN mkdir -p /app/image_cache && \
    chown -R appuser:appgroup /app

# Change to non-root user
USER appuser

# Expose port
EXPOSE 8080 8108

# Run the application
CMD ["./smlgoapi"]
