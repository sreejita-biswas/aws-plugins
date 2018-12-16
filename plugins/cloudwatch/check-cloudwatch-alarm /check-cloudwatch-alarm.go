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
# check-cloudwatch-alarm
#
# DESCRIPTION:
#   This plugin retrieves the state of a CloudWatch alarm. Can be configured
#   to trigger a warning or critical based on the result. Defaults to OK unless
#   alarm is missing
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
#
# USAGE:
#   ./check-cloudwatch-alarm --alarm_name=TestAlarm
#
# NOTES:
#
# LICENSE:
#   TODO
#
*/

var (
	alarmName        string
	criticalList     string
	warningList      string
	cloudWatchClient *cloudwatch.CloudWatch
	awsRegion        string
)

func main() {
	criticals := []string{}
	warnings := []string{}
	flag.StringVar(&alarmName, "alarm_name", "TestAlarm", "Alarm name")
	flag.StringVar(&criticalList, "criticals", "Alarm", "Comma seperated Critical List")
	flag.StringVar(&warningList, "warnings", "INSUFFICIENT_DATA", "Comma seperated Warning List")
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
	describeInput.AlarmNames = []*string{aws.String(alarmName)}

	describeOutput, err := cloudWatchClient.DescribeAlarms(describeInput)

	if err != nil {
		fmt.Println("Failed to get cloudwatch alarm details , Error : ", err)
	}

	if describeOutput == nil || describeOutput.MetricAlarms == nil || len(describeOutput.MetricAlarms) == 0 {
		fmt.Println("Unknown : Unable to find alarm")
		return
	}

	message := fmt.Sprintf("Alarm State : %s", *describeOutput.MetricAlarms[0].StateValue)

	criticals = strings.Split(criticalList, ",")

	for _, critical := range criticals {
		if *describeOutput.MetricAlarms[0].StateValue == strings.ToUpper(critical) {
			fmt.Println("CRITICAL :", message)
			return
		}
	}

	warnings = strings.Split(warningList, ",")
	for _, warning := range warnings {
		if *describeOutput.MetricAlarms[0].StateValue == strings.ToUpper(warning) {
			fmt.Println("WARNING :", message)
			return
		}
	}

	fmt.Println("OK :", message)
}
