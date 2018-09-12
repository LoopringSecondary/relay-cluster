#!/bin/bash
#ValidateService
WORK_DIR=/opt/loopring/extractor

#cron and logrotate are installed by default in ubuntu, don't check it again
if [ ! -f /etc/logrotate.d/loopring-extractor ]; then
    sudo cp $WORK_DIR/src/bin/logrotate/loopring-extractor /etc/logrotate.d/loopring-extractor
fi

pgrep cron
if [[ $? != 0 ]]; then
    sudo /etc/init.d/cron start
fi

#check later
exit 0
