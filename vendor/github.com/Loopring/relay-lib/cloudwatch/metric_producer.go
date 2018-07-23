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

package cloudwatch

import (
	"fmt"
	"github.com/Loopring/relay-lib/log"
	"github.com/Loopring/relay-lib/utils"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"time"
)

const namespace = "LoopringDefine"
const obsoleteCountThreshold = 200
const obsoleteTimeoutSeconds = 4
const batchDatumBufferSize = 400
const batchTimeoutSeconds = 2
const batchSendSize = 20 //aws only allow not more than 20 items in one request, request size should less than 40960

type CloudWatchConfig struct {
	Enabled bool
	Region string
}

type CloudWatchClient struct {
	innerClient *cloudwatch.CloudWatch
	enabled     bool
}

var cwc *CloudWatchClient

var inChan chan<- interface{}
var outChan <-chan interface{}

/*
 need following config files for aws service connect
	~/.aws/config/credentials

two ways to specify this config
1. export variable on start at /etc/server/xxx/run, when use daemontools
export AWS_SHARED_CREDENTIALS_FILE=/home/ubuntu/.aws/credentials
2. local run as current user, then will default use this credentials file base in home dir
*/

func Initialize(config CloudWatchConfig) error {
	if !config.Enabled {
		log.Infof("CloudWatch is not enabled")
		cwc = &CloudWatchClient{nil, false}
		return nil
	} else if len(config.Region) == 0 {
		log.Errorf("CloudWatchConfig.Region is empty")
		return fmt.Errorf("CloudWatchConfig.Region is empty")
	}

	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(config.Region),
		Credentials: credentials.NewSharedCredentials("", ""),
	})
	if err != nil {
		log.Errorf("Initialize cloudwatch metric producer failed : %s\n", err.Error())
		return err
	} else {
		cwc = &CloudWatchClient{cloudwatch.New(sess), true}
		inChan, outChan = utils.MakeInfinite()
		log.Info("Ready for produce cloudwatch metric\n")
		go func() {
			obsoleteCount := 0
			batchDatumBuffer := make([]*cloudwatch.MetricDatum, 0, batchDatumBufferSize)
			bufferStartTimeStamp := time.Now()
			for {
				select {
				case data, ok := <-outChan:
					if !ok {
						log.Error("Receive from cloud watch output channel failed")
					} else {
						datum, ok := data.(*cloudwatch.MetricDatum)
						if !ok {
							log.Error("Convert data to PutMetricDataInput failed")
						} else {
							if checkObsolete(datum) {
								obsoleteCount += 1
								if obsoleteCount >= obsoleteCountThreshold {
									log.Errorf("Obsolete cloud watch metric data count is %d, just drop\n", obsoleteCount)
									obsoleteCount = 0
								}
							} else {
								batchDatumBuffer = append(batchDatumBuffer, datum)
								if obsoleteCount > 0 {
									log.Errorf("Drop %d obsolete cloud watch metric data\n", obsoleteCount)
									obsoleteCount = 0
								}
							}
							if checkTimeout(datum, bufferStartTimeStamp) && len(batchDatumBuffer) > 0 || len(batchDatumBuffer) >= batchDatumBufferSize {
								batchSendMetricData(batchDatumBuffer)
								batchDatumBuffer = make([]*cloudwatch.MetricDatum, 0, batchDatumBufferSize)
								bufferStartTimeStamp = *datum.Timestamp
							}
						}
					}
				}
			}
		}()
		return nil
	}
}

func Close() {
	if IsValid() {
		close(inChan)
	}
}

func IsValid() bool {
	return cwc != nil && cwc.enabled
}

func PutResponseTimeMetric(methodName string, costTime float64) error {
	if !IsValid() {
		return nil
	}
	dt := &cloudwatch.MetricDatum{}
	metricName := fmt.Sprintf("response_%s", methodName)
	dt.MetricName = &metricName
	dt.Value = &costTime
	unit := cloudwatch.StandardUnitMilliseconds
	dt.Unit = &unit
	tms := time.Now()
	dt.Timestamp = &tms
	innerPutMetricData(dt)
	return nil
}

func PutHeartBeatMetric(metricName string) error {
	if !IsValid() {
		return nil
	}
	dt := &cloudwatch.MetricDatum{}
	dt.MetricName = &metricName
	heartbeatValue := 1.0
	dt.Value = &heartbeatValue
	unit := cloudwatch.StandardUnitCount
	dt.Unit = &unit
	tms := time.Now()
	dt.Timestamp = &tms
	innerPutMetricData(dt)
	return nil
}

func innerPutMetricData(datum *cloudwatch.MetricDatum) {
	// no dimension metric
	storeMetricLocal(datum)
	// host dimension metric
	cloneDatum := cloneDatum(datum)
	cloneDatum.Dimensions = []*cloudwatch.Dimension{}
	cloneDatum.Dimensions = append(cloneDatum.Dimensions, hostDimension())
	storeMetricLocal(cloneDatum)
}

func storeMetricLocal(datatum *cloudwatch.MetricDatum) error {
	inChan <- datatum
	return nil
}

func cloneDatum(datum *cloudwatch.MetricDatum) *cloudwatch.MetricDatum {
	dt := &cloudwatch.MetricDatum{}
	dt.MetricName = datum.MetricName
	dt.Value = datum.Value
	dt.Unit = datum.Unit
	dt.Timestamp = datum.Timestamp
	return dt
}

func batchSendMetricData(datums []*cloudwatch.MetricDatum) {
	//log.Infof("batchSendMetricData %s send datums size %d\n", time.Now().Format(time.RFC3339), len(datums))
	for i := 0; ; i++ {
		if i*batchSendSize >= len(datums) {
			return
		}
		input := &cloudwatch.PutMetricDataInput{}
		endIndex := (i + 1) * batchSendSize
		if endIndex > len(datums) {
			endIndex = len(datums)
		}
		input.MetricData = datums[i*batchSendSize : endIndex]
		input.Namespace = namespaceNormal()
		go func() {
			if _, err := cwc.innerClient.PutMetricData(input); err != nil {
				log.Errorf("cwc.PutMetricData failed with error : %s\n", err.Error())
			}
		}()
	}
}

func checkObsolete(datum *cloudwatch.MetricDatum) bool {
	//log.Infof("checkObsolete : %d %d %d \n", time.Now().UnixNano(), datum.Timestamp.UnixNano(), time.Now().UnixNano() - datum.Timestamp.UnixNano())
	return time.Now().UnixNano()-datum.Timestamp.UnixNano() > 1000*1000*1000*obsoleteTimeoutSeconds
}

func checkTimeout(datum *cloudwatch.MetricDatum, startTime time.Time) bool {
	return datum.Timestamp.UnixNano()-startTime.UnixNano() > 1000*1000*1000*batchTimeoutSeconds
}

func namespaceNormal() *string {
	sp := namespace
	return &sp
}

func hostDimension() *cloudwatch.Dimension {
	dim := &cloudwatch.Dimension{}
	ipDimName := "host"
	dim.Name = &ipDimName
	ip := utils.GetLocalIp()
	dim.Value = &ip
	return dim
}
