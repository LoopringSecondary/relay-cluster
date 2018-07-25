FROM golang:1.9-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers
RUN mkdir /opt /opt/loopring /opt/loopring/relay /opt/loopring/relay/logs /opt/loopring/relay/logs/motan /opt/loopring/relay/config

ENV WORKSPACE=$GOPATH/src/github.com/Loopring/relay-cluster
ADD . $WORKSPACE

RUN cd $WORKSPACE && go build -ldflags -s -v  -o build/bin/relay cmd/main.go
RUN mv $WORKSPACE/build/bin/relay /$GOPATH/bin

EXPOSE 8083 8087

ENTRYPOINT ["relay"]