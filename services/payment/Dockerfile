FROM golang:1.26.3-bookworm AS builder

RUN apt-get update && apt-get install -y git ca-certificates tzdata

WORKDIR /app/services/payment

COPY packages/go-shared /app/packages/go-shared
COPY services/payment/go.mod services/payment/go.sum ./
RUN go mod download

COPY services/payment/ .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" \
    -o /app/bin/payment \
    ./cmd/server

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

WORKDIR /app

COPY --from=builder /app/bin/payment /app/payment
COPY --from=builder /app/services/payment/migrations /app/migrations

EXPOSE 8080 9090

ENTRYPOINT ["/app/payment"]
