FROM golang:1.26-bookworm

RUN apt-get update && apt-get install -y gcc

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=1
RUN go build -o server ./cmd/server

EXPOSE 8080
CMD ["./server"]