#!/bin/sh
#ValidateService

WORK_DIR=/opt/loopring/relay

#cron and logrotate are installed by default in ubuntu, don't check it again
if [ ! -f /etc/logrotate.d/loopring-relay ]; then
    sudo cp $WORK_DIR/src/bin/logrotate/loopring-relay /etc/logrotate.d/loopring-relay
fi

pgrep cron
if [[ $? != 0 ]]; then
    sudo /etc/init.d/cron start
fi

k=1
RPC_PORT=8083
WAIT_SECONDS=120

echo "check rpc port......."
TEST_URL="127.0.0.1:$RPC_PORT"

for k in $(seq 1 $WAIT_SECONDS)
do
    sleep 1
    STATUS_CODE=`curl -o /dev/null -s -w %{http_code} $TEST_URL`
    if [ "$STATUS_CODE" = "200" ]; then
        echo "request test_url:$TEST_URL succeeded!"
        echo "response code:$STATUS_CODE"
        exit 0;
    else
        echo "request test_url:$TEST_URL failed!"
        echo "response code: $STATUS_CODE"
        echo "try one more time:the $k time....."
    fi
    if [ ${k} -eq ${WAIT_SECONDS} ]; then
        echo "have tried $k times, no more try"
        echo "failed"
        exit -1
    fi
done
