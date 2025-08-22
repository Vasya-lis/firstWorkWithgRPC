FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY .. .

RUN go mod download 

WORKDIR /app/cmd/db-service

RUN go build -o db-service


FROM alpine:3.18

WORKDIR /app

COPY --from=builder /app/cmd/db-service/db-service .
RUN chmod +x db-service

EXPOSE 50051

CMD ["./db-service"]