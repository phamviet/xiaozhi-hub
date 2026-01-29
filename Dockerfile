FROM --platform=$BUILDPLATFORM oven/bun:debian AS ui-builder

WORKDIR /app/ui

COPY ui/package.json ui/bun.lock ./
RUN bun install --no-save --frozen-lockfile

COPY ui/ ./
RUN bun run build

# ? -------------------------
FROM --platform=$BUILDPLATFORM golang:1.25 AS builder

WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source files
COPY . ./
COPY --from=ui-builder /app/ui/dist ./ui/dist

# Build
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=0 GOGC=75 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags "-w -s" -o pb .

# https://github.com/benbjohnson/litestream/blob/main/Dockerfile
FROM litestream/litestream

COPY --from=builder /app/pb /
COPY scripts/docker-entrypoint.sh /

# Ensure data persistence across container recreations
VOLUME ["/pb_data"]

EXPOSE 8090

ENTRYPOINT [ "/docker-entrypoint.sh" ]
CMD ["serve", "--http=0.0.0.0:8090"]