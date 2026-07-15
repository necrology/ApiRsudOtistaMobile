FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w" \
    -o /out/api \
    ./cmd/api

FROM alpine:3.22

RUN addgroup -S app && adduser -S app -G app \
    && apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=builder /out/api /app/api
COPY --from=builder /src/assets /app/assets

USER app

EXPOSE 8080

ENTRYPOINT ["/app/api"]
