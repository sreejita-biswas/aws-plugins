package main

/*
# check-ec2-network
#
# DESCRIPTION:
#   Check EC2 Network Metrics by CloudWatch API.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
# USAGE:
#   ./check-ec2-network.rb -r ${you_region} -i ${your_instance_id} --warning-over 1000000 --critical-over 1500000
#   ./check-ec2-network.rb -r ${you_region} -i ${your_instance_id} -d NetworkIn --warning-over 1000000 --critical-over 1500000
#   ./check-ec2-network.rb -r ${you_region} -i ${your_instance_id} -d NetworkOut --warning-over 1000000 --critical-over 1500000
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
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sreejita-biswas/aws-plugins/aws_clients"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
)

var (
	ec2Client         *ec2.EC2
	cloudWatchClient  *cloudwatch.CloudWatch
	criticalThreshold float64
	warningThreshold  float64
	instanceId        string
	endTime           string
	period            int64
	direction         string
)

func getEc2NetworkMetric(endTimeDate time.Time) (*float64, error) {
	stats := "Average"
	var input cloudwatch.GetMetricStatisticsInput
	input.Namespace = aws.String("AWS/EC2")
	input.MetricName = aws.String(direction)
	var dimensionFilter cloudwatch.Dimension
	dimensionFilter.Name = aws.String("InstanceId")
	dimensionFilter.Value = aws.String(instanceId)
	input.Dimensions = []*cloudwatch.Dimension{&dimensionFilter}
	input.EndTime = aws.Time(endTimeDate)
	input.StartTime = aws.Time(endTimeDate.Add(time.Duration(-15) * time.Minute))
	input.Period = aws.Int64(period)
	input.Statistics = []*string{aws.String(stats)}
	input.Unit = aws.String("Bytes")
	metrics, err := cloudWatchClient.GetMetricStatistics(&input)
	if err != nil {
		return nil, err
	}
	if metrics != nil && metrics.Datapoints != nil && len(metrics.Datapoints) > 1 {
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
		return averageValue, nil
	}
	return nil, nil
}

func main() {
	flag.Float64Var(&criticalThreshold, "critical", 1000000, "Trigger a critical if network traffice is over specified Bytes")
	flag.Float64Var(&warningThreshold, "warning", 1500000, "Trigger a warning if network traffice is over specified Bytes")
	flag.StringVar(&instanceId, "instance_id", "", "EC2 Instance ID to check.")
	flag.StringVar(&endTime, "start_time", time.Now().Format(time.RFC3339), "CloudWatch metric statistics end time, e.g. 2014-11-12T11:45:26.371Z")
	flag.Int64Var(&period, "period", 60, "CloudWatch metric statistics period in seconds")
	flag.StringVar(&direction, "direction", "NetworkIn", "Select NetworkIn or NetworkOut")
	flag.Parse()

	if instanceId == "" || len(strings.TrimSpace(instanceId)) == 0 {
		fmt.Println("Please enter a valid instance id.")
		return
	}

	if !(direction == "NetworkIn" || direction == "NetworkOut") {
		fmt.Println("Invalid direction")
		return
	}

	endTimeDate, err := time.Parse(time.RFC3339, endTime)

	if err != nil {
		fmt.Println("Invalid end time entered , ", err)
		return
	}

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

	networkValue, err := getEc2NetworkMetric(endTimeDate)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}
	if networkValue != nil {
		if *networkValue > criticalThreshold {
			fmt.Println("critical :", direction, "at", *networkValue, "Bytes")
		} else if *networkValue > warningThreshold {
			fmt.Println("warning :", direction, " at ", *networkValue, "Bytes")
		} else {
			fmt.Println("ok :", direction, "at", *networkValue, "Bytes")
		}
	}

}
