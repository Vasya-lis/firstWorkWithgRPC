FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY . .

RUN go mod download

WORKDIR /app/cmd/api

RUN go build -o api-service

FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/cmd/api/api-service .

COPY --from=builder /app/web ./web

EXPOSE 7540

CMD ["./api-service"]