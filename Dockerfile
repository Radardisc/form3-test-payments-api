FROM golang:latest
MAINTAINER Liam Galvin <liam@liam-galvin.co.uk>

RUN mkdir -p /go/src/github.com/liamg/form3-payments-api
WORKDIR /go/src/github.com/liamg/form3-payments-api
COPY . .

RUN go build -o /usr/bin/api .

EXPOSE 8080

ENTRYPOINT ["/usr/bin/api"]
