FROM golang:1.26.3-bookworm AS builder

RUN apt-get update && apt-get install -y git ca-certificates tzdata

WORKDIR /app/services/order

COPY services/order/go.mod services/order/go.sum ./
COPY services/order/vendor/ ./vendor/

COPY services/order/ .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -mod=vendor -ldflags="-w -s" \
    -o /app/bin/order \
    ./cmd/server

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

WORKDIR /app

COPY --from=builder /app/bin/order /app/order
COPY --from=builder /app/services/order/migrations /app/migrations

EXPOSE 8080 9090

ENTRYPOINT ["/app/order"]
