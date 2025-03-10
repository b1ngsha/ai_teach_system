FROM anolis-registry.cn-zhangjiakou.cr.aliyuncs.com/openanolis/golang:1.20.12-23

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

RUN go build -o sync ./cmd/sync

COPY entrypoint.sh .

RUN chmod +x entrypoint.sh

CMD ["./entrypoint.sh"]
