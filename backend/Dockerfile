# Build stage
FROM golang:1.22-alpine AS builder

# Set working directory
WORKDIR /app

# Install dependencies
# Using alpine-sdk for any potential CGO needs (though we'll try to disable it)
RUN apk add --no-cache git

# Copy go.mod and go.sum
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
# We target cmd/api/main.go as seen in the project structure
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-service ./cmd/api/main.go

# Final stage
FROM alpine:latest  

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from the builder stage
COPY --from=builder /app/api-service .
# Copy migrations to the path expected by the code (relative to WORKDIR)
COPY --from=builder /app/database/migrations ./database/migrations
# Copy .env.example as a base
COPY --from=builder /app/.env.example ./.env

# Expose port (matched with .env)
ENV PORT=3001
EXPOSE 3001

# Run the binary
CMD ["./api-service"]
