# STAGE 1: BUILD GOLANG BACKEND
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy backend's dependency files and download them.
COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./

# Build the go binary => creates a single, statically-linked executable
# CGO_ENABLED=0 is important for creating a truly portable binary
RUN CGO_ENABLED=0 go build -o /server .

# STAGE 2: BUILD TYPESCRIPT FRONTEND
FROM node:20-alpine AS frontend-builder

WORKDIR /app

COPY frontend/ ./

RUN npm install -g pnpm
RUN pnpm install
RUN pnpm exec tsc

# STAGE 3: FINAL PRODUCTION IMAGE
FROM gcr.io/distroless/static-debian11 AS final

WORKDIR /app

# Copy the compiled Go binary from builder stage
COPY --from=builder /server /app/server

# Copy the compiled frontend assets from the 'frontend-builder' stage
COPY --from=frontend-builder /app/index.html /app/frontend/index.html
COPY --from=frontend-builder /app/style.css /app/frontend/style.css
COPY --from=frontend-builder /app/dist/main.js /app/frontend/dist/main.js

# Expose the port our server listens on
EXPOSE 8080

# The command to run when the container starts
CMD ["/app/server"]
