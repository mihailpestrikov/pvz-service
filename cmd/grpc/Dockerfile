FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o grpc ./cmd/grpc

FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/grpc .

EXPOSE 3000
EXPOSE 9000

CMD ["./grpc"]