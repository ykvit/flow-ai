# =================================================================
# Stage 1: Frontend Builder
# =================================================================
FROM node:20-alpine AS builder-frontend
WORKDIR /app
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# =================================================================
# Stage 2: Backend Builder
# =================================================================
FROM golang:1.22-alpine AS builder-backend
WORKDIR /src
RUN apk add --no-cache build-base
RUN addgroup -S appgroup -g 1001 && adduser -S appuser -u 1001 -G appgroup
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
# FINAL FIX: Restore the missing build command!
# The output binary will be created at /src/server
RUN go build -ldflags="-w -s" -o ./server ./cmd/server
RUN chown -R appuser:appgroup /src

# =================================================================
# Stage 3: Docs Generator
# =================================================================
FROM golang:1.22-alpine AS docs-generator
WORKDIR /src
RUN go install github.com/swaggo/swag/cmd/swag@latest
COPY backend/ ./
RUN swag init -g cmd/server/main.go

# =================================================================
# Stage 4: Final Production Image
# =================================================================
FROM alpine:3.20 AS final
WORKDIR /app
RUN apk add --no-cache nginx ca-certificates
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
COPY --from=builder-backend /src/server /usr/local/bin/server
COPY --from=builder-frontend /app/dist /app/frontend/dist
COPY --from=docs-generator /src/docs /app/docs
COPY nginx.conf /etc/nginx/nginx.conf
COPY --chmod=755 docker/entrypoint.sh /entrypoint.sh
RUN chown -R appuser:appgroup /app && \
    mkdir -p /data /var/log/nginx /var/lib/nginx/tmp && \
    chown -R appuser:appgroup /data /var/log/nginx /var/lib/nginx
USER appuser
EXPOSE 80
ENTRYPOINT ["/entrypoint.sh"]