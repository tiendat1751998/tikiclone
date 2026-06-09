FROM golang:1.26.3-bookworm AS builder

RUN apt-get update && apt-get install -y git ca-certificates tzdata

WORKDIR /app/services/notification-campaign

COPY packages/go-shared /app/packages/go-shared
COPY services/notification-campaign/go.mod services/notification-campaign/go.sum ./
RUN go mod download

COPY services/notification-campaign/ .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" \
    -o /app/bin/notification \
    ./cmd/server

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

WORKDIR /app

COPY --from=builder /app/bin/notification /app/notification
COPY --from=builder /app/services/notification-campaign/migrations /app/migrations

EXPOSE 8087 9097

ENTRYPOINT ["/app/notification"]