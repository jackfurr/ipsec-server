FROM golang:1.8.3-alpine

ENV webserver_path /go/src/github.com/jackfurr/ipsec-server/
ENV PATH $PATH:$webserver_path

WORKDIR $webserver_path
COPY . .

RUN apk -Uuv add openssl git
RUN sh "$(pwd)/ssh-cert.sh"
# RUN go get -u
RUN go get github.com/tkanos/gonfig
RUN go get github.com/go-sql-driver/mysql
RUN go build .

ENTRYPOINT ./main

EXPOSE 80 443