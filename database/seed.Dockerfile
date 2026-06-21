FROM golang:1.26-bookworm

RUN apt-get update && apt-get install -y gcc

WORKDIR /seed

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY database/seed.go .
COPY database/sample-data.json .

ENV CGO_ENABLED=1
RUN go build -o seed seed.go

CMD ["./seed", "sample-data.json"]