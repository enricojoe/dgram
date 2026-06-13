# syntax=docker/dockerfile:1

# ---- Build stage ----
FROM golang:1.26-alpine AS build
WORKDIR /src

# Cache dependencies first.
COPY go.mod go.sum ./
RUN go mod download

# Build the static server binary. Migrations are embedded, so no extra files
# need to be copied into the runtime image.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/server ./cmd/server

# ---- Runtime stage ----
FROM alpine:3.20 AS runtime
WORKDIR /app

# Non-root user and CA certs (TLS-capable, e.g. for outbound calls).
RUN apk add --no-cache ca-certificates \
 && adduser -D -u 10001 app
USER app

COPY --from=build /out/server /app/server

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/server"]
