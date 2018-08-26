#!/bin/bash
#AfterInstall

WORK_DIR=/opt/loopring/extractor
SVC_DIR=/etc/service/extractor
GOROOT=/usr/lib/go-1.9
export PATH=$PATH:$GOROOT/bin
export GOPATH=/opt/loopring/go-src

#cp svc config to svc if this node is not miner
sudo cp -rf $WORK_DIR/src/bin/svc/* $SVC_DIR
sudo chmod -R 755 $SVC_DIR

SRC_DIR=$GOPATH/src/github.com/Loopring/extractor
if [ ! -d $SRC_DIR ]; then
      sudo mkdir -p $SRC_DIR
	  sudo chown -R ubuntu:ubuntu $GOPATH
fi

cd $SRC_DIR
rm -rf ./*
cp -r $WORK_DIR/src/* ./
go build -ldflags -s -v  -o build/bin/extractor cmd/main.go
cp build/bin/extractor $WORK_DIR/bin/
