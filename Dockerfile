FROM oven/bun:1.3.6-debian AS ui-builder

WORKDIR /app/ui

COPY ui/package.json ui/bun.lock ./
RUN bun install --no-save --frozen-lockfile

COPY ui/ ./
RUN bun run build

# ? -------------------------
FROM --platform=$BUILDPLATFORM golang:1.25 AS builder
RUN DEBIAN_FRONTEND=noninteractive \
    apt-get update && \
    apt-get install -y libopus-dev libopusfile-dev libsoxr-dev

WORKDIR /app

RUN wget https://github.com/k2-fsa/sherpa-onnx/releases/download/asr-models/silero_vad.onnx

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source files
COPY . ./
COPY --from=ui-builder /app/ui/dist ./ui/dist

# Build
ARG TARGETOS TARGETARCH
RUN CGO_ENABLED=1 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -ldflags "-w -s" -o pb .

# https://github.com/benbjohnson/litestream/blob/main/Dockerfile
FROM litestream/litestream

RUN apt-get update && apt-get install -y --no-install-recommends \
    libopus0 \
    libopusfile0 \
    libsoxr0 \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir -p /pb_data /models

COPY --from=builder /app/pb /
COPY --from=builder /app/silero_vad.onnx /models/
COPY --from=builder /go/pkg/mod/github.com/k2-fsa/sherpa-onnx-go-linux@v1.12.23/lib/aarch64-unknown-linux-gnu/libsherpa-onnx-c-api.so /lib/aarch64-linux-gnu/
COPY --from=builder /go/pkg/mod/github.com/k2-fsa/sherpa-onnx-go-linux@v1.12.23/lib/aarch64-unknown-linux-gnu/libonnxruntime.so /lib/aarch64-linux-gnu/

COPY scripts/docker-entrypoint.sh /

VOLUME ["/pb_data"]

EXPOSE 8090

ENTRYPOINT [ "/docker-entrypoint.sh" ]
CMD ["serve", "--http=0.0.0.0:8090"]