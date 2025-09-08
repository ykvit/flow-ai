# =================================================================
# Target: 'builder-frontend' - Builds the React static files
# =================================================================
FROM node:20-alpine AS builder-frontend
WORKDIR /app

# Copy package files and install dependencies
COPY frontend/package.json frontend/package-lock.json* ./
RUN npm install

# Copy the rest of the frontend source code and build it
COPY frontend/ ./
RUN npm run build


# =================================================================
# Target: 'builder-backend' - Compiles the Go application
# =================================================================
FROM golang:1.22-alpine AS builder-backend
WORKDIR /src

RUN apk add --no-cache build-base

# Copy Go module files
COPY backend/go.mod backend/go.sum ./

# It makes the build process more robust.
RUN go mod tidy

# Download dependencies based on the now-correct go.sum
RUN go mod download

# Copy the rest of the backend source code
COPY backend/ ./

RUN go build -ldflags="-w -s" -o /app/server ./cmd/server


# =================================================================
# Target: 'frontend' (Final Production Image for NGINX)
# =================================================================
FROM nginx:1.25-alpine AS frontend

# Copy the built static files from the frontend builder stage
COPY --from=builder-frontend /app/dist /usr/share/nginx/html

# The NGINX config will be mounted via compose.yaml
# Expose port 80 for NGINX
EXPOSE 80

# The default NGINX command runs automatically
CMD ["nginx", "-g", "daemon off;"]


# =================================================================
# Target: 'backend' (Final Production Image for Go)
# =================================================================
FROM alpine:latest AS backend
WORKDIR /app

# Copy just the compiled Go binary from the backend builder stage
COPY --from=builder-backend /app/server .

# Expose the port the Go app runs on
EXPOSE 8000

# The command to run the application
CMD ["./server"]