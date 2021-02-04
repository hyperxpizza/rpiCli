FROM golang:1.15.6-alpine

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh openssl

WORKDIR /app

COPY go.mod go.sum ./

RUN pwd
RUN ls -la

RUN go mod download

COPY . .

RUN go build -o server server/server.go

#RUN ./server/cert/gen.sh

EXPOSE 9999

CMD ["./server"]
