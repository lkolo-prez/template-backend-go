# syntax=docker/dockerfile:1.9
# ---- Multi-stage build for a static, distroless, non-root Go service. ----

# Pin the builder to a digest in real usage; tag shown for readability.
FROM golang:1.26-bookworm AS build

# Build metadata injected via --build-arg and stamped into the binary.
ARG VERSION=dev
ARG COMMIT=none
ARG BUILD_DATE=unknown
ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /src

# Cache modules separately from source for faster incremental builds.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

COPY . .

# CGO disabled -> fully static binary; trimpath + ldflags strip and stamp.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build \
      -trimpath \
      -buildvcs=false \
      -ldflags="-s -w \
        -X main.version=${VERSION} \
        -X main.commit=${COMMIT} \
        -X main.date=${BUILD_DATE}" \
      -o /out/server ./cmd/server

# ---- Final stage: distroless static, non-root, read-only friendly. ----
FROM gcr.io/distroless/static-debian12:nonroot AS final

# OCI labels for provenance and discoverability.
LABEL org.opencontainers.image.title="template-backend-go" \
      org.opencontainers.image.description="Production-grade Go HTTP microservice template" \
      org.opencontainers.image.source="https://github.com/aircloud/template-backend-go" \
      org.opencontainers.image.licenses="Apache-2.0" \
      org.opencontainers.image.vendor="aircloud"

WORKDIR /
COPY --from=build /out/server /server

# distroless:nonroot already runs as uid 65532; declare it explicitly.
USER 65532:65532

EXPOSE 8080
ENV HOST=0.0.0.0 PORT=8080

# Liveness probe using the binary itself (no shell available in distroless).
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/server", "-healthcheck"]

ENTRYPOINT ["/server"]
