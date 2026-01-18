FROM --platform=$BUILDPLATFORM oven/bun:alpine AS ui-builder

WORKDIR /app/ui

COPY ui/package.json ui/bun.lock ./
RUN bun install --no-save --frozen-lockfile

COPY ui/ ./
RUN bun run build

# ? -------------------------
FROM --platform=$BUILDPLATFORM golang:alpine AS builder

WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source files
COPY . ./
COPY --from=ui-builder /app/ui/dist ./ui/dist

RUN apk add --no-cache \
    unzip \
    ca-certificates

RUN update-ca-certificates

# Build
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOGC=75 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags "-w -s" -o pb .

# ? -------------------------
FROM alpine

RUN apk add --no-cache \
    ca-certificates

COPY --from=builder /app/pb /

# Ensure data persistence across container recreations
VOLUME ["/pb_data"]

EXPOSE 8090

ENTRYPOINT [ "/pb" ]
CMD ["serve", "--http=0.0.0.0:8090"]