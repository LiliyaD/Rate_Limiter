FROM golang:alpine

ADD . /src
RUN cd /src && go build -o server cmd/main.go

ENTRYPOINT ["/src/server"]