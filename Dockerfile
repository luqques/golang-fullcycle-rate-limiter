FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod ./
COPY . .

RUN go mod tidy

RUN go build -trimpath -ldflags="-s -w" -o /rate-limiter ./cmd/api

FROM alpine:3.20

RUN adduser -D -H appuser

USER appuser

COPY --from=builder /rate-limiter /rate-limiter

EXPOSE 8080

ENTRYPOINT ["/rate-limiter"]