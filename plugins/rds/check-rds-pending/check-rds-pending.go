package main

/*
#
# check-rds-pending
#
#
# DESCRIPTION:
#   This plugin checks rds clusters for pending maintenance action.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
#
# USAGE:
#  ./check-rds-pending --aws_region=${you_region}
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

	"github.com/sreejita-biswas/aws-plugins/awsclient"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
)

var (
	awsRegion string
	rdsClient *rds.RDS
)

func main() {
	var success bool
	flag.StringVar(&awsRegion, "aws_region", "us-east-1", "AWS Region (defaults to us-east-1).")
	flag.Parse()
	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)
	success, rdsClient = awsclient.GetRDSClient(awsSession)
	if !success {
		return
	}
	clusters, err := getClusters()
	if err != nil || clusters == nil {
		return
	}
	checkPendingMaintenance(clusters)
}

func getClusters() ([]*string, error) {
	clusters := []*string{}
	dbInstanceInput := &rds.DescribeDBInstancesInput{}
	//fetch all clusters identifiers
	dbClusterOutput, err := rdsClient.DescribeDBInstances(dbInstanceInput)

	if err != nil {
		fmt.Println("An error occurred processing AWS RDS API DescribeDBInstances", err)
		return nil, err
	}

	if dbClusterOutput != nil && dbClusterOutput.DBInstances != nil && len(dbClusterOutput.DBInstances) > 0 {
		for _, dbInstance := range dbClusterOutput.DBInstances {
			clusters = append(clusters, dbInstance.DBInstanceIdentifier)
		}
	}

	if !(clusters != nil && len(clusters) > 0) {
		fmt.Println("OK")
		return nil, nil
	}

	return clusters, nil
}

func checkPendingMaintenance(clusters []*string) {
	pendingMaintanceInput := &rds.DescribePendingMaintenanceActionsInput{}
	filter := &rds.Filter{}
	filter.Name = aws.String("db-instance-id")
	filter.Values = clusters
	pendingMaintanceInput.Filters = []*rds.Filter{filter}
	pendingMaintanceOutput, err := rdsClient.DescribePendingMaintenanceActions(pendingMaintanceInput)
	if err != nil {
		fmt.Println("Error :", err)
		return
	}

	if pendingMaintanceOutput == nil || pendingMaintanceOutput.PendingMaintenanceActions == nil || len(pendingMaintanceOutput.PendingMaintenanceActions) == 0 {
		return
	}

	fmt.Println("CRITICAL : Clusters w/ pending maintenance required:")
	for _, pendingMaintance := range pendingMaintanceOutput.PendingMaintenanceActions {
		fmt.Println(pendingMaintance.PendingMaintenanceActionDetails)
	}
}
