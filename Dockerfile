FROM golang:1.23-alpine

# Set environment variables
ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY src/ ./src/

WORKDIR /app/src
RUN go build -o /app/main

EXPOSE 8088

CMD ["/app/main"]
