FROM golang:1.8.3-alpine

ENV webserver_path /go/src/gitlab.com/accounts4/ipsec-server/
ENV PATH $PATH:$webserver_path

WORKDIR $webserver_path
COPY . .

RUN apk -Uuv add openssl
RUN sh "$(pwd)/install.sh"
RUN go get -u
RUN go build .

ENTRYPOINT ./main

EXPOSE 80 443