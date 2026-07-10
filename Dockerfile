# ─────────────────────────────────────────────────────────────────────────────
# Stage 1 — builder
#
# Pakai debian-based golang image (bukan alpine) karena CGO butuh glibc +
# libopus yang tersedia via apt. Alpine pakai musl libc dan pkg opus-nya
# lebih ribet — lebih aman pakai bookworm-slim.
# ─────────────────────────────────────────────────────────────────────────────
FROM golang:1.25-bookworm AS builder

WORKDIR /app

# Install libopus dev headers dulu — dibutuhkan gopkg.in/hraban/opus.v2 (CGO).
RUN apt-get update && apt-get install -y --no-install-recommends \
    pkg-config \
    libopus-dev \
    libopusfile-dev \
    && rm -rf /var/lib/apt/lists/*

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=1 wajib untuk opus. Binary akan dynamically linked ke libopus
# (ditangani di runtime stage di bawah).
RUN CGO_ENABLED=1 GOOS=linux go build -trimpath -ldflags="-s -w" -o leviathan ./cmd

# ─────────────────────────────────────────────────────────────────────────────
# Stage 2 — runtime
#
# Pakai debian-slim (bukan alpine) agar shared library libopus.so tersedia
# dan ABI-nya cocok dengan binary yang dikompilasi di stage builder.
# ─────────────────────────────────────────────────────────────────────────────
FROM debian:bookworm-slim

# ca-certificates : TLS calls ke LLM provider (Gemini/OpenRouter).
# wget            : dipakai HEALTHCHECK di bawah.
# libopus0        : shared library runtime untuk gopkg.in/hraban/opus.v2.
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    wget \
    libopus0 \
    libopusfile0 \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd -r leviathan && useradd -r -g leviathan leviathan

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
