package main

/*
# s3-metrics
#
# DESCRIPTION:
#   Gets S3 metrics from CloudWatch and puts them in Graphite for longer term storage
#
# OUTPUT:
#   metric-data
#
# PLATFORMS:
#   MAC OS
#
# USAGE:
#   ./metrics-s3
#
# NOTES:
#
# LICENSE:
#   TODO
#
*/

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sreejita-biswas/aws-plugins/aws_clients"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
)

var (
	s3Client         *s3.S3
	scheme           string
	awsRegion        string
	cloudWatchClient *cloudwatch.CloudWatch
)

func main() {
	flag.StringVar(&awsRegion, "aws_region", "us-east-2", "AWS Region (defaults to us-east-1).")
	flag.StringVar(&scheme, "scheme", "sensu.aws.s3.buckets", "Metric naming scheme, text to prepend to metric")
	flag.Parse()

	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)

	if awsSession != nil {
		s3Client = aws_clients.NewS3(awsSession)
	} else {
		fmt.Println("Error while getting aws session")
		os.Exit(0)
	}

	if s3Client == nil {
		fmt.Println("Error while getting s3 client session")
		os.Exit(0)
	}

	cloudWatchClient = aws_clients.NewCloudWatch(awsSession)

	if cloudWatchClient == nil {
		fmt.Errorf("Failed to create cloud watch client")
	}

	input := &s3.ListBucketsInput{}
	output, err := s3Client.ListBuckets(input)
	if err != nil {
		fmt.Println(err)
		return
	}

	if output != nil && output.Buckets != nil && len(output.Buckets) > 1 {
		for _, bucket := range output.Buckets {
			bucketName := strings.Replace(*bucket.Name, ".", "_", -1)
			getMetricStatistics(bucketName)
		}
	}

}

func getMetricStatistics(bucketName string) {
	stats := "Average"
	var period int64
	period = 24 * 60 * 60
	var input cloudwatch.GetMetricStatisticsInput
	input.Namespace = aws.String("AWS/S3")
	input.MetricName = aws.String("BucketSizeBytes")
	var dimensionFilter cloudwatch.Dimension
	dimensionFilter.Name = aws.String("BucketName")
	dimensionFilter.Value = aws.String(bucketName)
	var dimensionFilter2 cloudwatch.Dimension
	dimensionFilter2.Name = aws.String("StorageType")
	dimensionFilter2.Value = aws.String("StandardStorage")
	input.Dimensions = []*cloudwatch.Dimension{&dimensionFilter, &dimensionFilter2}
	input.EndTime = aws.Time(time.Now())
	input.StartTime = aws.Time(time.Now().Add(time.Duration(-24*60) * time.Minute))
	input.Period = aws.Int64(period)
	input.Statistics = []*string{aws.String(stats)}
	input.Unit = aws.String("Count")
	metrics, err := cloudWatchClient.GetMetricStatistics(&input)
	if err != nil {
		fmt.Println("CRITICAL :", scheme, ".", bucketName, ".", "Error : ", err)
	}
	if metrics != nil {
		var minimumTimeDifference float64
		var timeDifference float64
		var averageValue *float64
		minimumTimeDifference = -1
		for _, datapoint := range metrics.Datapoints {
			timeDifference = time.Since(*datapoint.Timestamp).Seconds()
			if minimumTimeDifference == -1 {
				minimumTimeDifference = timeDifference
				averageValue = datapoint.Average
			} else if timeDifference < minimumTimeDifference {
				minimumTimeDifference = timeDifference
				averageValue = datapoint.Average
			}
		}
		fmt.Println(scheme, ".", bucketName, ".", "Number of objects :", averageValue)
	}
}
