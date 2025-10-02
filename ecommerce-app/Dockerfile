# Stage 1: Build the application
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum to download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code and migrations
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ecommerce-app ./main.go

# Stage 2: Create a lean runtime image
FROM alpine:3.20

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/ecommerce-app .
# Copy the migrations directory
COPY --from=builder /app/migrations ./migrations

# Expose the port
EXPOSE 8080

# Default command (overridden by docker-compose for worker)
CMD ["./ecommerce-app"]
