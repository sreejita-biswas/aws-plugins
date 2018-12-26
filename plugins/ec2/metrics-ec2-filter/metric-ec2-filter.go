package main

/*
#
# check-ec2-filter
#
# DESCRIPTION:
#   This plugin retrieves EC2 instances matching a given filter and
#   returns the number matched. Warning and Critical thresholds may be set as needed.
#   Thresholds may be compared to the count using [equal, not, greater, less] operators.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
#
# USAGE:
#   ./metric-ec2-filter -filters="{\"filters\" : [{\"name\" : \"instance-state-name\", \"values\": [\"running\"]}]}"
# NOTES:
#
# LICENSE:
#   TODO
#
*/

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
	"github.com/sreejita-biswas/aws-plugins/awsclient"
	"github.com/sreejita-biswas/aws-plugins/models"
	"github.com/sreejita-biswas/aws-plugins/utils"
)

var (
	ec2Client  *ec2.EC2
	filters    string
	metricType string
	scheme     string
	filterName string
)

func main() {
	var success bool
	flag.StringVar(&metricType, "metric_type", "instance", "Count by type: status, instance")
	flag.StringVar(&scheme, "scheme", "sensu.aws.ec2", "Metric naming scheme, text to prepend to metric")
	flag.StringVar(&filters, "filters", "{}", "JSON String representation of Filters, e.g. {\"filters\" : [{\"name\" : \"instance-state-name\", \"values\": [\"running\"]}]}")
	flag.StringVar(&filterName, "filter_name", "", "Filter naming scheme, text to prepend to metric")
	flag.Parse()

	awsSession := aws_session.CreateAwsSession()

	success, ec2Client = awsclient.GetEC2Client(awsSession)
	if !success {
		return
	}
	var ec2Fileters models.Filters
	err := json.Unmarshal([]byte(filters), &ec2Fileters)
	if err != nil {
		fmt.Println("Failed to unmarshal filter data , ", err)
	}

	reservations, err := utils.GetReservations(ec2Client, ec2Fileters.Filters)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	fmt.Print("EC2 Instances of ", scheme)
	if len(strings.TrimSpace(filterName)) > 1 {
		fmt.Println("with filter :", filterName)
	} else {
		fmt.Println()
	}
	for _, reservation := range reservations {
		for _, instance := range reservation.Instances {
			fmt.Println(*instance.InstanceId)
		}
	}

}
