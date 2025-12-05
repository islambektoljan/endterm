FROM golang:1.21-alpine

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY *.go .

RUN go build -o app .

EXPOSE 8080

CMD ["./app"]