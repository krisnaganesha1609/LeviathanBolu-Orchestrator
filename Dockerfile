FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# CGO_ENABLED=0 keeps the binary statically linked so it runs unmodified
# on the minimal alpine runtime image below.
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o leviathan ./cmd

FROM alpine:latest

# ca-certificates: needed for any outbound TLS calls (LLM provider APIs).
# wget: used by the HEALTHCHECK below to hit /livez.
RUN apk --no-cache add ca-certificates wget \
    && addgroup -S leviathan && adduser -S leviathan -G leviathan

WORKDIR /home/leviathan

COPY --from=builder /app/leviathan .

# NOTE: no `COPY .env .` here on purpose. Baking a .env file into the image
# bakes in whoever built it secrets and makes the image stop being
# portable/rebuildable by anyone else. docker-compose.yml injects the
# actual environment at *run* time via `env_file`, which is the right
# layer for secrets to live in.

USER leviathan

EXPOSE 8009

HEALTHCHECK --interval=15s --timeout=3s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8009/livez || exit 1

CMD ["./leviathan"]
