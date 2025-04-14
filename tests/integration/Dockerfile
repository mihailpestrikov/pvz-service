FROM golang:1.24-alpine

WORKDIR /app

RUN apk add --no-cache curl

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go get github.com/stretchr/testify/require

COPY ./tests/integration/wait-for-api.sh /wait-for-api.sh
RUN chmod +x /wait-for-api.sh

CMD ["/wait-for-api.sh", "go", "test", "-v", "./tests/integration/..."]