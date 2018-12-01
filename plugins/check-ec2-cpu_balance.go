package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sreejita-biswas/aws-plugins/aws_clients"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
	"github.com/sreejita-biswas/aws-plugins/config"
)

var (
	conf             config.Config
	ec2Client        *ec2.EC2
	cloudWatchClient *cloudwatch.CloudWatch
)

func getEc2CpuBalance(instance ec2.Instance) (*float64, error) {
	stats := "Average"
	var period int64
	period = 60
	var input cloudwatch.GetMetricStatisticsInput
	input.Namespace = aws.String("AWS/EC2")
	input.MetricName = aws.String("CPUCreditBalance")
	var dimensionFilter cloudwatch.Dimension
	dimensionFilter.Name = aws.String("InstanceId")
	dimensionFilter.Value = instance.InstanceId
	input.Dimensions = []*cloudwatch.Dimension{&dimensionFilter}
	input.StartTime = aws.Time(time.Now())
	input.EndTime = aws.Time(time.Now())
	input.Period = aws.Int64(period)
	input.Statistics = []*string{aws.String(stats)}
	metrics, err := cloudWatchClient.GetMetricStatistics(&input)
	if err != nil {
		return nil, err
	}
	if metrics != nil {
		for _, datapoint := range metrics.Datapoints {
			return datapoint.Average, nil
		}
	}
	return nil, nil
}

func getMatchingInstanceTag(instance ec2.Instance) *string {

	for _, tag := range instance.Tags {
		if *tag.Key == conf.Tag {
			return tag.Value
		}
	}
	return nil

}

func getReservations() ([]*ec2.Reservation, error) {
	filter := ec2.Filter{Name: aws.String("instance-state-name"), Values: []*string{
		aws.String("running")}}

	input := &ec2.DescribeInstancesInput{
		Filters: []*ec2.Filter{&filter},
	}

	result, err := ec2Client.DescribeInstances(input)
	if err != nil {
		return nil, err
	}

	return result.Reservations, nil
}

func main() {

	var reservations []*ec2.Reservation
	os.Setenv("AWS_ACCESS_KEY_ID", "AKIAIOSFODNN7EXAMPLE")
	os.Setenv("AWS_ACCESS_KEY", "AKIAIOSFODNN7EXAMPLE")
	os.Setenv("AWS_SECRET_ACCESS_KEY", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")
	os.Setenv("AWS_SECRET_KEY", "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY")
	os.Setenv("AWS_REGION", "us-west-2")
	os.Setenv("AWS_DEFAULT_REGION", "us-west-2")

	awsSession := aws_session.CreateAwsSession()

	if awsSession != nil {
		ec2Client = aws_clients.NewEC2(awsSession)
	} else {
		fmt.Errorf("Error while getting aws session")
		os.Exit(0)
	}

	if ec2Client == nil {
		fmt.Errorf("Error while getting ec2 client session")
		os.Exit(0)
	}

	reservations, err := getReservations()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	cloudWatchClient = aws_clients.NewCloudWatch()

	if cloudWatchClient == nil {
		fmt.Errorf("Failed to create cloud watch client")
	}

	for _, reservation := range reservations {
		for _, instance := range reservation.Instances {
			instanceType := instance.InstanceType
			if strings.HasPrefix(*instanceType, "t2.") {
				cpuBalance, err := getEc2CpuBalance(*instance)
				if err != nil {
					fmt.Errorf(err.Error())
					os.Exit(0)
				}
				tagValue := getMatchingInstanceTag(*instance)
				if tagValue != nil {
					if *cpuBalance < conf.Critical {
						fmt.Println(*instance.InstanceId, *tagValue, "is below critical threshold", " [ cpuBalance < ", conf.Critical, " ]")
					} else if *cpuBalance < conf.Warning {
						fmt.Println(*instance.InstanceId, *tagValue, "is below warning threshold", " [ cpuBalance < ", conf.Warning, " ]")
					}
				}
			}
		}
	}
}
