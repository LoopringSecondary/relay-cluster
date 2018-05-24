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
 need following local config files for aws service connect
	~/.aws/config/config
	~/.aws/config/credentials
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
