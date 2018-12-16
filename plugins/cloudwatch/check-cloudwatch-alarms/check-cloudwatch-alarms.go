package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/sreejita-biswas/aws-plugins/aws_clients"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
)

/*
#
# check-cloudwatch-alarms
#
# DESCRIPTION:
#   This plugin raise a critical if one of cloud watch alarms are in given state.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
#
# USAGE:
#   ./check-cloudwatch-alarms --exclude_alarms=CPUAlarmLow
#   ./check-cloudwatch-alarms --aws_region=eu-west-1 --exclude_alarms=CPUAlarmLow
#   ./check-cloudwatch-alarms --state=ALEARM
#
# NOTES:
#
# LICENSE:
#   TODO
#
*/

var (
	excludeAlarms    string
	state            string
	cloudWatchClient *cloudwatch.CloudWatch
	awsRegion        string
)

func main() {
	selectedAlarms := []string{}
	excludeAlarmsMap := make(map[string]*string)
	discardedAlarms := []string{}
	flag.StringVar(&excludeAlarms, "exclude_alarms", "", "Exclude alarms")
	flag.StringVar(&state, "state", "ALARM", "State of the alarm")
	flag.StringVar(&awsRegion, "aws_region", "us-east-2", "AWS Region (defaults to us-east-1).")
	flag.Parse()

	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)

	if awsSession != nil {
		cloudWatchClient = aws_clients.NewCloudWatch(awsSession)
	} else {
		fmt.Println("Error while getting aws session")
		os.Exit(0)
	}

	if cloudWatchClient == nil {
		fmt.Errorf("Failed to create cloud watch client")
		return
	}

	describeInput := &cloudwatch.DescribeAlarmsInput{}
	describeInput.StateValue = aws.String(state)

	describeOutput, err := cloudWatchClient.DescribeAlarms(describeInput)

	if err != nil {
		fmt.Println("Failed to get cloudwatch alarm details , Error : ", err)
	}

	if describeOutput == nil || describeOutput.MetricAlarms == nil || len(describeOutput.MetricAlarms) == 0 {
		fmt.Println("OK : No alarm in", state, "state")
		return
	}

	if len(excludeAlarms) > 0 {
		discardedAlarms = strings.Split(excludeAlarms, ",")
		for _, alarm := range discardedAlarms {
			excludeAlarmsMap[alarm] = &alarm
		}
	}

	for _, alarm := range describeOutput.MetricAlarms {
		if excludeAlarmsMap[*alarm.AlarmName] == nil {
			selectedAlarms = append(selectedAlarms, *alarm.AlarmName)
		}
	}

	if selectedAlarms != nil && len(selectedAlarms) > 0 {
		fmt.Println("CRITICAL :", len(selectedAlarms), "are in state", state, " :", selectedAlarms)
		return
	}

	fmt.Println("OK : Everything looks good")
}
