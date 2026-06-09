FROM golang:1.26.3-bookworm AS builder

RUN apt-get update && apt-get install -y git ca-certificates tzdata

WORKDIR /app/services/oms-fulfillment

COPY packages/go-shared /app/packages/go-shared
COPY services/oms-fulfillment/go.mod services/oms-fulfillment/go.sum ./
RUN go mod download

COPY services/oms-fulfillment/ .

RUN go mod tidy && go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" \
    -o /app/bin/oms-fulfillment \
    ./cmd/server

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

WORKDIR /app

COPY --from=builder /app/bin/oms-fulfillment /app/oms-fulfillment

EXPOSE 8095 9095

ENTRYPOINT ["/app/oms-fulfillment"]
