# syntax=docker/dockerfile:1
# Multi-stage build for the cart service (build context = repo root).
FROM golang:1.23-alpine AS builder
WORKDIR /src
ENV CGO_ENABLED=0 GOOS=linux GOFLAGS=-mod=mod
COPY go.mod ./
COPY pkg ./pkg
COPY services/cart ./services/cart
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go build -trimpath -ldflags="-s -w" -o /out/cart ./services/cart/cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /
COPY --from=builder /out/cart /cart
EXPOSE 8080
USER nonroot:nonroot
HEALTHCHECK --interval=10s --timeout=3s --start-period=20s --retries=3 \
    CMD ["/cart", "healthcheck"]
ENTRYPOINT ["/cart"]
