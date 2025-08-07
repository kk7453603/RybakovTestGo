FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

RUN adduser -D -s /bin/sh -u 1001 appuser

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download
RUN go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o server ./cmd/server

FROM scratch

COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

COPY --from=builder /app/server /server

USER appuser

EXPOSE 8080 9090

ENTRYPOINT ["/server"]
