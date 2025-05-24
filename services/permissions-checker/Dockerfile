# Stage 1: Builder
# Use a Go base image with Alpine for a smaller build environment
FROM golang:1.24-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker's build cache
# This means `go mod download` will only run if dependencies change
COPY go.mod go.sum ./

# Download Go modules
# `go mod download` is generally preferred over `go mod tidy` in Dockerfiles
# because it only fetches necessary dependencies.
# If you have private modules, you might need to add authentication here.
RUN go mod download

# Copy the rest of your application's source code
# This includes `main.go`, `config/`, `cache/`, `proto/`, and the generated `auth/` directory
COPY . .

# Build the Go application
# CGO_ENABLED=0 for static compilation (no C dependencies, good for Alpine/scratch)
# GOOS=linux ensures it's built for a Linux environment
# -a -installsuffix nocgo helps ensure a fully static binary
# -o auth_service specifies the output binary name
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o auth_service .

# Stage 2: Runner
# Use a minimal Alpine base image for the final production image
FROM alpine:latest

# Set the working directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/auth_service .

# Copy the Casbin model configuration file
COPY --from=builder /app/model.conf .

# Install ca-certificates for HTTPS/TLS connections
# This is crucial if your NATS, Valkey, or other external services use TLS
RUN apk add --no-cache ca-certificates

# Expose the gRPC port
EXPOSE 50051

# Command to run the application
# This will be the entry point when the container starts
CMD ["./auth_service"]
