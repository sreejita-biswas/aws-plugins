package main

/*
# metrics-ec2-count
#
# DESCRIPTION:
#   This plugin retrieves number of EC2 instances.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
# USAGE:
#   # get metrics on the status of all instances in the region
#   ./metrics-ec2-count.go --metric_type=status
#
#   # get metrics on all instance types in the region
#   ./metrics-ec2-count.go --metric_type=instance
#
# NOTES:
#
# LICENSE:
#  TODO
*/

import (
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sreejita-biswas/aws-plugins/aws_clients"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
	"github.com/sreejita-biswas/aws-plugins/utils"
)

var (
	ec2Client        *ec2.EC2
	cloudWatchClient *cloudwatch.CloudWatch
	metricType       string
	scheme           string
)

func main() {
	flag.StringVar(&metricType, "metric_type", "instance", "Count by type: status, instance")
	flag.StringVar(&scheme, "scheme", "sensu.aws.ec2", "Metric naming scheme, text to prepend to metric")
	flag.Parse()

	metricCount := make(map[string]int)

	awsSession := aws_session.CreateAwsSession()

	if awsSession != nil {
		ec2Client = aws_clients.NewEC2(awsSession)
	} else {
		fmt.Println("Error while getting aws session")
		os.Exit(0)
	}

	if ec2Client == nil {
		fmt.Println("Error while getting ec2 client session")
		os.Exit(0)
	}

	cloudWatchClient = aws_clients.NewCloudWatch(awsSession)

	if cloudWatchClient == nil {
		fmt.Println("Failed to create cloud watch client")
	}

	reservations, err := utils.GetReservations(ec2Client, nil)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	for _, reservation := range reservations {
		for _, instance := range reservation.Instances {
			if metricType == "status" {
				metricCount[*instance.State.Name] = metricCount[*instance.State.Name] + 1
			}
			if metricType == "instance" {
				metricCount[*instance.InstanceType] = metricCount[*instance.InstanceType] + 1
			}
		}
	}

	fmt.Println("Number of", scheme, "instances by", metricType)
	for metric, count := range metricCount {
		fmt.Println(metric, "-", count)
	}
}
