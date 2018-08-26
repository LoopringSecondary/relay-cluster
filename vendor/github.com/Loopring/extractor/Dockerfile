FROM golang:1.9-alpine as builder

RUN apk add --no-cache make gcc musl-dev linux-headers
RUN mkdir /opt /opt/loopring /opt/loopring/extractor /opt/loopring/extractor/logs /opt/loopring/extractor/config

ENV WORKSPACE=$GOPATH/src/github.com/Loopring/extractor
ADD . $WORKSPACE

RUN cd $WORKSPACE && go build -ldflags -s -v  -o build/bin/extractor cmd/main.go
RUN mv $WORKSPACE/build/bin/extractor /$GOPATH/bin

ENTRYPOINT ["extractor"]