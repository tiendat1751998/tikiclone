FROM golang:1.26.3-bookworm AS builder

RUN apt-get update && apt-get install -y git ca-certificates tzdata

WORKDIR /app/services/cart

COPY packages/go-shared /app/packages/go-shared
COPY services/cart/go.mod services/cart/go.sum ./
COPY services/cart/ .
RUN go mod tidy

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" \
    -o /app/bin/cart \
    ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

WORKDIR /app

COPY --from=builder /app/bin/cart /app/cart
COPY --from=builder /app/services/cart/migrations /app/migrations

EXPOSE 8080 9090

ENTRYPOINT ["/app/cart"]
