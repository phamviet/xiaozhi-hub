FROM oven/bun:1.3.6-debian AS ui-builder

WORKDIR /app/ui

COPY ui/package.json ui/bun.lock ./
RUN bun install --no-save --frozen-lockfile

COPY ui/ ./
RUN bun run build

# ? -------------------------
FROM golang:1.25 AS builder
RUN apt-get update && apt-get install -y --no-install-recommends \
    libopus-dev libopusfile-dev libsoxr-dev \
    && rm -rf /var/lib/apt/lists/*

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
RUN CGO_ENABLED=1 go build -ldflags "-w -s" -o pb .

RUN if [ "$TARGETARCH" = "arm64" ]; then \
        LD=aarch64-unknown-linux-gnu; \
    else \
      LD=x86_64-unknown-linux-gnu; \
    fi && \
    SHERPA_ONNX_VERSION=$(go list -m -f '{{ .Version }}' github.com/k2-fsa/sherpa-onnx-go-linux) && \
    SHERPA_ONNX_LIB_PATH=/go/pkg/mod/github.com/k2-fsa/sherpa-onnx-go-linux@${SHERPA_ONNX_VERSION}/lib/${LD} && \
    mkdir -p /app/lib && \
    cd $SHERPA_ONNX_LIB_PATH && \
    cp libsherpa-onnx-c-api.so libonnxruntime.so /app/lib/

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
COPY --from=builder /app/lib/* /lib/

COPY scripts/docker-entrypoint.sh /

VOLUME ["/pb_data"]

EXPOSE 8090

ENTRYPOINT [ "/docker-entrypoint.sh" ]
CMD ["serve", "--http=0.0.0.0:8090"]