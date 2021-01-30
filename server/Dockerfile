FROM golang:1.15.6-alpine

RUN apk update && apk upgrade && \
    apk add --no-cache bash git openssh openssl

COPY ./../go.mod ./
COPY ./../go.sum ./

RUN go mod download

COPY . .

RUN go build -o main .

RUN cd cert/ && ./gen.sh

EXPOSE 9999

CMD ["./main"]