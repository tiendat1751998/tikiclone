FROM golang:1.26.3-bookworm AS builder

RUN apt-get update && apt-get install -y git ca-certificates tzdata

WORKDIR /app/services/payment-ledger

COPY packages/go-shared /app/packages/go-shared
COPY services/payment-ledger/go.mod services/payment-ledger/go.sum ./
RUN go mod download

COPY services/payment-ledger/ .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" \
    -o /app/bin/payment-ledger \
    ./cmd/server

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

WORKDIR /app

COPY --from=builder /app/bin/payment-ledger /app/payment-ledger

EXPOSE 8097 9097

ENTRYPOINT ["/app/payment-ledger"]
