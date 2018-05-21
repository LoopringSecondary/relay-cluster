package watchcloud

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"net"
	"time"
)

const region = "ap-northeast-1"
const namespace = "LoopringDefine"

var cwc *cloudwatch.CloudWatch

func Initialize() error {
	//NOTE: use default config ~/.asw/credentials
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(region),
		Credentials: credentials.NewSharedCredentials("", ""),
	})
	if err != nil {
		return err
	} else {
		cwc = cloudwatch.New(sess)
		return nil
	}
}

func IsValid() bool {
	return cwc != nil
}

func PutResponseTimeMetric(methodName string, costTime float64) error {
	if cwc != nil {
		return fmt.Errorf("Cloudwatch client has not initialized\n")
	}
	dt := &cloudwatch.MetricDatum{}
	metricName := fmt.Sprintf("response_%s", methodName)
	dt.MetricName = &metricName
	dt.Dimensions = []*cloudwatch.Dimension{}
	dt.Dimensions = append(dt.Dimensions, hostDimension())
	dt.Value = &costTime
	unit := cloudwatch.StandardUnitMilliseconds
	dt.Unit = &unit
	tms := time.Now()
	dt.Timestamp = &tms

	datums := []*cloudwatch.MetricDatum{}
	datums = append(datums, dt)

	input := &cloudwatch.PutMetricDataInput{}
	input.MetricData = datums
	input.Namespace = namespaceNormal()

	_, err := cwc.PutMetricData(input)
	return err
}

func PutHeartBeatMetric(metricName string) error {
	if cwc != nil {
		return fmt.Errorf("Cloudwatch client has not initialized\n")
	}
	dt := &cloudwatch.MetricDatum{}
	dt.MetricName = &metricName
	dt.Dimensions = []*cloudwatch.Dimension{}
	dt.Dimensions = append(dt.Dimensions, globalDimension())
	hearbeatValue := 1.0
	dt.Value = &hearbeatValue
	unit := cloudwatch.StandardUnitCount
	dt.Unit = &unit
	tms := time.Now()
	dt.Timestamp = &tms

	datums := []*cloudwatch.MetricDatum{}
	datums = append(datums, dt)

	input := &cloudwatch.PutMetricDataInput{}
	input.MetricData = datums
	input.Namespace = namespaceNormal()

	_, err := cwc.PutMetricData(input)
	return err
}

func namespaceNormal() *string {
	sp := namespace
	return &sp
}

func hostDimension() *cloudwatch.Dimension {
	dim := &cloudwatch.Dimension{}
	ipDimName := "host"
	dim.Name = &ipDimName
	dim.Value = getIp()
	return dim
}

func globalDimension() *cloudwatch.Dimension {
	dim := &cloudwatch.Dimension{}
	dimName := "global"
	dim.Name = &dimName
	dimValue := "nt"
	dim.Value = &dimValue
	return dim
}

func getIp() *string {
	var res string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		res = "unknown"
	}
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				res = ipnet.IP.To4().String()
			}
		}
	}
	res = "unknown"
	return &res
}
