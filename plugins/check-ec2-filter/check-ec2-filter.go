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
#   ./check-ec2-filter -filters="{\"filters\" : [{\"name\" : \"instance-state-name\", \"values\": [\"running\"]}]}"
#   ./check-ec2-filter -exclude_tags="{\"TAG_NAME\" : \"TAG_VALUE\"}" -compare=not
# NOTES:
#
# LICENSE:
#   Justin McCarty
#   Released under the same terms as Sensu (the MIT license); see LICENSE
#   for details.
#
*/

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sreejita-biswas/aws-plugins/aws_clients"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
	"github.com/sreejita-biswas/aws-plugins/models"
	"github.com/sreejita-biswas/aws-plugins/utils"
)

var (
	ec2Client               *ec2.EC2
	cloudWatchClient        *cloudwatch.CloudWatch
	criticalThreshold       int
	warningThreshold        int
	excludeTags             string
	compareValue            string
	detailedMessageRequired bool
	minRunningSecs          float64
	filters                 string
)

func main() {

	flag.IntVar(&criticalThreshold, "critical", 1, "Critical threshold for filter")
	flag.IntVar(&warningThreshold, "warning", 2, "Warning threshold for filter',	")
	flag.StringVar(&excludeTags, "exclude_tags", "{}", "JSON String Representation of tag values")
	flag.StringVar(&compareValue, "compare", "equal", "Comparision operator for threshold: equal, not, greater, less")
	flag.BoolVar(&detailedMessageRequired, "detailed_message", false, "Detailed description is required or not")
	flag.Float64Var(&minRunningSecs, "min_running_secs", 0, "Minimum running seconds")
	flag.StringVar(&filters, "filters", "{\"filters\" : [{\"name\" : \"instance-state-name\", \"values\": [\"running\"]}]}", "JSON String representation of Filters")
	flag.Parse()

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

	var excludedTags map[string]*string
	err := json.Unmarshal([]byte(excludeTags), &excludedTags)
	if err != nil {
		fmt.Println("Failed to unmarshal exclude tags details , ", err)
	}

	var ec2Fileters models.Filters
	err = json.Unmarshal([]byte(filters), &ec2Fileters)
	if err != nil {
		fmt.Println("Failed to unmarshal filter data , ", err)
	}

	reservations, err := utils.GetReservations(ec2Client, ec2Fileters.Filters)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(0)
	}

	var awsInstances []models.AwsInstance
	for _, reservation := range reservations {
		for _, instance := range reservation.Instances {
			tags := instance.Tags
			excludeIntance := false
			for _, tag := range tags {
				if excludedTags[*tag.Key] != nil && *excludedTags[*tag.Key] == *tag.Value {
					excludeIntance = true
				}
			}
			if !excludeIntance {
				timeDifference := time.Since(time.Now().Add(time.Duration(-10) * time.Minute)).Seconds()
				if !(timeDifference < minRunningSecs) {
					awsInstance := models.AwsInstance{Id: *instance.InstanceId, LaunchTime: *instance.LaunchTime, Tags: instance.Tags}
					awsInstances = append(awsInstances, awsInstance)
				}
			}
		}
	}

	selectedInstancesCount := len(awsInstances)
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("Current Count : %d  ", selectedInstancesCount))
	if detailedMessageRequired && selectedInstancesCount > 0 {
		for _, awsInstance := range awsInstances {
			buffer.WriteString(fmt.Sprintf(", %s", awsInstance.Id))
		}
	}

	if compareValue == "equal" {
		if selectedInstancesCount == criticalThreshold {
			fmt.Println("Critical threshold for filter , ", buffer.String())
		}
		if selectedInstancesCount == warningThreshold {
			fmt.Println("Warning threshold for filter , ", buffer.String())
		}
	} else if compareValue == "not" {
		if selectedInstancesCount != criticalThreshold {
			fmt.Println("Critical threshold for filter , ", buffer.String())
		}
		if selectedInstancesCount != warningThreshold {
			fmt.Println("Warning threshold for filter , ", buffer.String())
		}
	} else if compareValue == "greater" {
		if selectedInstancesCount > criticalThreshold {
			fmt.Println("Critical threshold for filter , ", buffer.String())
		}
		if selectedInstancesCount > warningThreshold {
			fmt.Println("Warning threshold for filter , ", buffer.String())
		}
	} else if compareValue == "less" {
		if selectedInstancesCount < criticalThreshold {
			fmt.Println("Critical threshold for filter , ", buffer.String())
		}
		if selectedInstancesCount < warningThreshold {
			fmt.Println("Warning threshold for filter , ", buffer.String())
		}
	}
}
