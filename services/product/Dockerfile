# syntax=docker/dockerfile:1
#
# Multi-stage build for the product catalog service. Identical in shape to the
# user service's Dockerfile (build context = repo root so it can see go.mod and
# the shared pkg/ libraries).

# ---- Stage 1: build -------------------------------------------------------
FROM golang:1.23-alpine AS builder
WORKDIR /src
ENV CGO_ENABLED=0 GOOS=linux GOFLAGS=-mod=mod

COPY go.mod ./
COPY pkg ./pkg
COPY services/product ./services/product

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/product ./services/product/cmd/server

# ---- Stage 2: runtime -----------------------------------------------------
FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=builder /out/product /product

EXPOSE 8080
USER nonroot:nonroot

HEALTHCHECK --interval=10s --timeout=3s --start-period=20s --retries=3 \
    CMD ["/product", "healthcheck"]

ENTRYPOINT ["/product"]
