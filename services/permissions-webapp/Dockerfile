# Stage 1: Build the application
FROM --platform=$BUILDPLATFORM golang:1.24.2-alpine AS builder

WORKDIR /app

# Copy Go modules and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire application source code
COPY . .

# Ensure templ CLI is available and generate Go code from .templ files
# We can install templ CLI or assume it's run before docker build and .templ.go files are present.
# For a self-contained Docker build, installing it is safer.
RUN go install github.com/a-h/templ/cmd/templ@latest
RUN templ generate

# Build the Go application
# CGO_ENABLED=0 for static linking (good for alpine)
# -ldflags="-w -s" to strip debug symbols and reduce binary size
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags="-w -s" -o /permissions-webapp ./main.go

# Stage 2: Create the final lightweight image
FROM alpine:3.19

WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /permissions-webapp /app/permissions-webapp

# Copy static assets (like htmx.min.js)
# Ensure your static files are in ./internal/static relative to Dockerfile context
COPY ./internal/static ./static

# The Casbin model file will be mounted via ConfigMap in Kubernetes,
# so we don't need to copy it here.

# Expose the port the application listens on (defined by PORT env var, default 3001)
# This is informational; actual port mapping is done in Kubernetes Service.
EXPOSE 3001

# Set the entrypoint for the container
ENTRYPOINT ["/app/permissions-webapp"]
