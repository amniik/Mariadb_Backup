FROM golang:1.21.1-alpine3.18 as builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64


RUN mkdir -p /app
WORKDIR /app

COPY . .
RUN go mod tidy \
    && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./pkg/cmd/main.go

FROM mariadb:10.3.8
COPY --from=builder /app/main /app/pkg/backup/mysqldump.sh /app/main /app/pkg/backup/mariabackup.sh ./
CMD ["./main"]
