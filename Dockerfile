FROM golang:1.22-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/main ./cmd/server

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/main /app/main

EXPOSE 3333

CMD ["/app/main"]