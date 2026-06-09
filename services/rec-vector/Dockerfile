FROM golang:1.26.3-bookworm AS builder

RUN apt-get update && apt-get install -y git ca-certificates tzdata

WORKDIR /app/services/rec-vector

COPY packages/go-shared /app/packages/go-shared
COPY services/rec-vector/go.mod services/rec-vector/go.sum ./
RUN go mod download

COPY services/rec-vector/ .

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-w -s" \
    -o /app/bin/rec-vector \
    ./cmd/server

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

WORKDIR /app

COPY --from=builder /app/bin/rec-vector /app/rec-vector

EXPOSE 8096 9096

ENTRYPOINT ["/app/rec-vector"]
