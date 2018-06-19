/*

  Copyright 2017 Loopring Project Ltd (Loopring Foundation).

  Licensed under the Apache License, Version 2.0 (the "License");
  you may not use this file except in compliance with the License.
  You may obtain a copy of the License at

  http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing, software
  distributed under the License is distributed on an "AS IS" BASIS,
  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
  See the License for the specific language governing permissions and
  limitations under the License.

*/

package sns

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	"time"
)

/*
 need following config files for aws service connect
	~/.aws/config/credentials

two ways to specify this config
1. export variable on start at /etc/server/xxx/run, when use daemontools
export AWS_SHARED_CREDENTIALS_FILE=/home/ubuntu/.aws/credentials
2. local run as current user, then will default use this credentials file base in home dir
*/

type SnsClient struct {
	innerClient *sns.SNS
	topicArn    string
	valid       bool
}

type SnsConfig struct {
	SNSTopicArn string
}

const region = "ap-northeast-1"

var sc *SnsClient

func Initialize(config SnsConfig) (*SnsClient, error) {
	if len(config.SNSTopicArn) == 0 {
		return nil, fmt.Errorf("Sns TopicArn not set, will not init sns client")
	}
	//NOTE: use default config ~/.asw/credentials
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewSharedCredentials("", ""),
	})
	if err != nil {
		return nil, err
	} else {
		sc = &SnsClient{sns.New(sess), config.SNSTopicArn, true}
		return sc, nil
	}
}

func PublishSns(subject string, message string) error {
	if sc == nil {
		return fmt.Errorf("SnsClient not initialized, will not send message")
	} else {
		input := &sns.PublishInput{}
		input.SetTopicArn(sc.topicArn)
		input.SetSubject(subject)
		input.SetMessage(fmt.Sprintf("%s|%s", time.Now().Format("15:04:05"), message))
		_, err := sc.innerClient.Publish(input)
		if err != nil {
			return fmt.Errorf("Failed send sns message with error : %s\nSubject: %s\n, Message %s\n", err.Error(), subject, message)
		} else {
			return nil
		}
	}
}
