FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o kotoha ./cmd/server

FROM alpine:3.21
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/kotoha .
COPY --from=builder /app/internal/i18n/locales ./internal/i18n/locales
EXPOSE 8080
CMD ["./kotoha"]
