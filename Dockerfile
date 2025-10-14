# =================================================================
# Stage 1: Frontend Builder
# Builds the static frontend assets.
# =================================================================
FROM node:20-alpine AS builder-frontend
WORKDIR /app
# Copy package files first to leverage Docker layer caching for dependencies.
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# =================================================================
# Stage 2: Backend Builder & Developer's Toolbox
# A comprehensive environment for building, testing, and developing the Go backend.
# =================================================================
# Using a Debian-based image (bookworm) for robust CGo support, specifically for go-sqlite3.
FROM golang:1.24.8-bookworm AS builder-backend
WORKDIR /src

# Install essential build tools. build-essential provides gcc for CGo.
RUN apt-get update && apt-get install -y build-essential git wget
# Create a non-root user for security best practices.
RUN addgroup --system --gid 1001 appgroup && adduser --system --uid 1001 --ingroup appgroup appuser

# Pin all development tool versions for fully reproducible builds.
ENV MIGRATE_VERSION=v4.17.1
# Build `migrate` from source with the 'sqlite3' tag to embed the necessary CGo driver.
# This ensures the CLI works correctly without runtime dependencies.
RUN go install -tags 'sqlite3' github.com/golang-migrate/migrate/v4/cmd/migrate@${MIGRATE_VERSION} && \
    mv /go/bin/migrate /usr/local/bin/migrate
RUN go install golang.org/x/tools/cmd/goimports@v0.37.0
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.5.0
RUN go install github.com/swaggo/swag/cmd/swag@v1.16.6

# Download Go modules before copying the rest of the source code to leverage Docker layer caching.
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
# Build a statically linked, stripped binary for a smaller final image.
RUN go build -ldflags="-w -s" -o ./server ./cmd/server
# Ensure the non-root user owns the source and compiled binary.
RUN chown -R appuser:appgroup /src

# =================================================================
# Stage 3: Final Production Image
# A minimal, secure image for production deployment.
# =================================================================
FROM alpine:3.22 AS final
WORKDIR /app
# Install only necessary runtime dependencies: nginx for proxying and ca-certificates for TLS.
RUN apk add --no-cache nginx ca-certificates
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Copy only the essential compiled artifacts from previous stages.
COPY --from=builder-backend /src/server /usr/local/bin/server
COPY --from=builder-frontend /app/dist /app/frontend/dist
# Migrations are needed at runtime for the application to manage the database schema.
COPY backend/internal/database/migrations /app/migrations

COPY nginx.conf /etc/nginx/nginx.conf
COPY --chmod=755 docker/entrypoint.sh /entrypoint.sh
# Create and set permissions for the data volume and Nginx log/pid directories.
RUN chown -R appuser:appgroup /app && \
    mkdir -p /data /var/log/nginx /var/lib/nginx/tmp && \
    chown -R appuser:appgroup /data /var/log/nginx /var/lib/nginx

USER appuser
EXPOSE 80
ENTRYPOINT ["/entrypoint.sh"]