# services/permissions-editor/Dockerfile
# Stage 1: Builder
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Ensure templ CLI is available and generate Go code from .templ files
# We can install templ CLI or assume it's run before docker build and .templ.go files are present.
# For a self-contained Docker build, installing it is safer.
RUN go install github.com/a-h/templ/cmd/templ@latest

COPY go.mod go.sum ./
COPY vendor/ vendor/
# RUN go mod download

COPY . .


RUN templ generate

# IMPORTANT: Ensure your Go module name matches the Dockerfile output name
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix nocgo -o permissions-editor .

# Stage 2: Runner
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/permissions-editor .
COPY --from=builder /app/static/ /app/static/

RUN apk add --no-cache ca-certificates

EXPOSE 8080

CMD ["./permissions-editor"]
