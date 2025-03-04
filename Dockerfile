FROM golang:1.23.3

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

RUN go build -o sync cmd/sync

COPY entrypoint.sh .

CMD ["./entrypoint.sh"]
