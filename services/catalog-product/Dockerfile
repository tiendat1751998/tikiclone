FROM golang:1.26.3-bookworm AS builder

RUN apt-get update && apt-get install -y git ca-certificates tzdata

WORKDIR /app/services/catalog-product

COPY packages/go-shared /app/packages/go-shared
COPY proto /app/proto
COPY services/catalog-product/go.mod services/catalog-product/go.sum ./
RUN go mod download

COPY services/catalog-product/ .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" \
    -o /app/bin/catalog-product \
    ./cmd/server

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /app/bin/catalog-product /app/catalog-product
COPY --from=builder /app/services/catalog-product/migrations /app/migrations

WORKDIR /app

EXPOSE 8080 9090

ENTRYPOINT ["/app/catalog-product"]
