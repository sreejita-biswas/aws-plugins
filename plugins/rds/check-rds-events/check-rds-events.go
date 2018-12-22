package main

/*
#
# check-rds-events
#
#
# DESCRIPTION:
#   This plugin checks rds clusters for critical events.
#   Due to the number of events types on RDS clusters, the check
#   should filter out non-disruptive events that are part of
#   basic operations.
#
#   More info on RDS events:
#   http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_Events.html
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
#
# USAGE:
#  Check's a specific RDS instance in a specific region for critical events
#  ./check-rds-events --aws_region=${your_region}  --db_instance_id=${your_rds_instance_id_name}
#
#  Checks all RDS instances in a specific region
#  ./check-rds-events.rb --aws_region=${your_region}
#
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
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/sreejita-biswas/aws-plugins/aws_clients"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
)

var (
	awsRegion    string
	dbInstanceId string
	ec2Client    *ec2.EC2
	rdsClient    *rds.RDS
)

func main() {

	flag.StringVar(&awsRegion, "aws_region", "us-east-1", "AWS Region (defaults to us-east-1).")
	flag.StringVar(&dbInstanceId, "db_instance_id", "", "DB instance identifier")
	flag.Parse()

	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)

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

	resultRegions, err := ec2Client.DescribeRegions(nil)
	if err != nil {
		fmt.Println("Error", err)
		return
	}

	validRegion := false
	clusters := []string{}
	criticalClusters := []string{}

	if resultRegions != nil && resultRegions.Regions != nil && len(resultRegions.Regions) > 0 {
		for _, region := range resultRegions.Regions {
			if *region.RegionName == awsRegion {
				validRegion = true
				break
			}
		}
	}

	if !validRegion {
		fmt.Println("CRITICAL : Invalid region specified!")
		return
	}

	rdsClient = aws_clients.NewRDS(awsSession)

	if rdsClient == nil {
		fmt.Println("Error while getting rds client session")
		os.Exit(0)
	}

	dbInstanceInput := &rds.DescribeDBInstancesInput{}
	filter := &rds.Filter{}
	if len(dbInstanceId) > 0 {
		filter.Name = aws.String("db-instance-id")
		filter.Values = []*string{aws.String(dbInstanceId)}
		dbInstanceInput.Filters = []*rds.Filter{filter}
	}

	dbClusterOutput, err := rdsClient.DescribeDBInstances(dbInstanceInput)

	if err != nil {
		fmt.Println("An error occurred processing AWS RDS API DescribeDBInstances", err)
	}

	if dbClusterOutput != nil && dbClusterOutput.DBInstances != nil && len(dbClusterOutput.DBInstances) > 0 {
		clusters = append(clusters, dbInstanceId)
	} else {
		fmt.Println("UNKNOWN :", dbInstanceId, "instance not found")
		return
	}

	if len(dbInstanceId) == 0 {
		filter = &rds.Filter{}
		dbInstanceInput.Filters = []*rds.Filter{filter}
		//fetch all clusters identifiers
		dbClusterOutput, err = rdsClient.DescribeDBInstances(dbInstanceInput)
	}

	if err != nil {
		fmt.Println("An error occurred processing AWS RDS API DescribeDBInstances", err)
		return
	}

	if dbClusterOutput != nil && dbClusterOutput.DBInstances != nil && len(dbClusterOutput.DBInstances) > 0 {
		for _, dbInstance := range dbClusterOutput.DBInstances {
			clusters = append(clusters, *dbInstance.DBInstanceIdentifier)
		}
	}

	// fetch the last 15 minutes of events for each cluster
	//that way, we're only spammed with persistent notifications that we'd care about.

	if !(clusters != nil && len(clusters) > 0) {
		return
	}

	for _, cluster := range clusters {
		eventInput := &rds.DescribeEventsInput{}
		eventInput.SourceType = aws.String("DBInstance")
		eventInput.SourceIdentifier = &cluster
		eventInput.StartTime = aws.Time(time.Now().Add(time.Duration(-15) * time.Minute))
		eventOutput, err := rdsClient.DescribeEvents(eventInput)

		if err != nil {
			fmt.Println("Error occurred while getting rds event details for db instance -", cluster)
			return
		}

		if eventOutput == nil || eventOutput.Events == nil || len(eventOutput.Events) == 0 {
			continue
		}

		for _, event := range eventOutput.Events {
			// we will need to filter out non-disruptive/basic operation events.
			//ie. the regular backup operations
			if *event.Message == "//Backing up DB instance|Finished DB Instance backup|Restored from snapshot//" {
				continue
			}
			// ie. Replication resumed
			if *event.Message == "//Replication for the Read Replica resumed//" {
				continue
			}

			// you can add more filters to skip more events.

			// draft the messages
			criticalClusters = append(criticalClusters, fmt.Sprintf("%s : %s", cluster, *event.Message))
		}
	}

	if len(criticalClusters) > 0 {
		fmt.Println("CRITICAL : Clusters w/ critical events :", criticalClusters)
	}
}
