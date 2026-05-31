# syntax=docker/dockerfile:1
#
# Multi-stage build for the user service.
#   * Build context is the REPO ROOT (see docker-compose.yml) so the build can
#     see both go.mod and the shared pkg/ libraries.
#   * Stage 1 compiles a fully static binary with the Go toolchain.
#   * Stage 2 is Google's "distroless" static image: no shell, no package
#     manager, runs as a non-root user — a tiny (~10 MB) and hardened runtime.

# ---- Stage 1: build -------------------------------------------------------
FROM golang:1.23-alpine AS builder
WORKDIR /src

# Static, reproducible-ish build. GOFLAGS=-mod=mod lets the build resolve and
# write go.sum on the fly; once you run `go mod tidy` locally and commit go.sum,
# you can drop -mod=mod and add a cached `go mod download` layer for speed.
ENV CGO_ENABLED=0 GOOS=linux GOFLAGS=-mod=mod

COPY go.mod ./
COPY pkg ./pkg
COPY services/user ./services/user

# BuildKit cache mounts keep the module and build caches warm between builds.
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/user ./services/user/cmd/server

# ---- Stage 2: runtime -----------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=builder /out/user /user

EXPOSE 8080
USER nonroot:nonroot

# The binary probes itself (distroless has no curl/wget). See main.go.
HEALTHCHECK --interval=10s --timeout=3s --start-period=20s --retries=3 \
    CMD ["/user", "healthcheck"]

ENTRYPOINT ["/user"]
