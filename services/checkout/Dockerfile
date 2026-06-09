FROM golang:1.26.3-bookworm AS builder

RUN apt-get update && apt-get install -y git ca-certificates tzdata

WORKDIR /app/services/checkout

COPY packages/go-shared /app/packages/go-shared
COPY services/checkout/go.mod services/checkout/go.sum ./
RUN go mod download

COPY services/checkout/ .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" \
    -o /app/bin/checkout \
    ./cmd/server

FROM scratch

COPY --from=builder /app/bin/checkout /app/checkout
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

WORKDIR /app

EXPOSE 8080 9090

ENTRYPOINT ["/app/checkout"]
