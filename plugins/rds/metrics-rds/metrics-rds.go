package main

/*
#
# rds-metrics
#
# DESCRIPTION:
#   Gets RDS metrics from CloudWatch and puts them in Graphite for longer term storage
#
# OUTPUT:
#   metric-data
#
# PLATFORMS:
#   MAC OS
#
#
# USAGE:
#   ./rds-metrics --aws_region=eu-west-1
#   ./rds-metrics --aws_region=eu-west-1 --db_instance_id=sr2x8pbti0eon1
#
# NOTES:
#   Returns all RDS statistics for all RDS instances in this account unless you specify --db_instance_id
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
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/sreejita-biswas/aws-plugins/aws_clients"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
)

var (
	awsRegion        string
	rdsClient        *rds.RDS
	scheme           string
	dbInstanceId     string
	fetchAge         int
	period           int64
	statistics       string
	cloudWatchClient *cloudwatch.CloudWatch
)

func main() {

	statisticType := "Average"
	clusters := []*string{}
	statisticsTypeMap := make(map[string]string)
	statisticsTypeMap["CPUUtilization"] = statisticType
	statisticsTypeMap["DatabaseConnections"] = statisticType
	statisticsTypeMap["FreeStorageSpace"] = statisticType
	statisticsTypeMap["ReadIOPS"] = statisticType
	statisticsTypeMap["ReadLatency"] = statisticType
	statisticsTypeMap["ReadThroughput"] = statisticType
	statisticsTypeMap["WriteIOPS"] = statisticType
	statisticsTypeMap["WriteLatency"] = statisticType
	statisticsTypeMap["WriteThroughput"] = statisticType
	statisticsTypeMap["ReplicaLag"] = statisticType
	statisticsTypeMap["SwapUsage"] = statisticType
	statisticsTypeMap["BinLogDiskUsage"] = statisticType
	statisticsTypeMap["DiskQueueDepth"] = statisticType

	flag.StringVar(&awsRegion, "aws_region", "us-west-1", "AWS Region (defaults to us-east-1).")
	flag.StringVar(&scheme, "scheme", "", "Metric naming scheme, text to prepend to metric")
	flag.StringVar(&dbInstanceId, "db_instance_id", "", "DB instance identifier")
	flag.IntVar(&fetchAge, "fetch_age", 0, "How long ago to fetch metrics from in seconds")
	flag.Int64Var(&period, "period", 60, "CloudWatch metric statistics period")
	flag.StringVar(&statistics, "statistics", "average", "CloudWatch statistics method")
	flag.Parse()

	if len(dbInstanceId) > 0 {
		clusters = []*string{&dbInstanceId}
	}

	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)

	if awsSession != nil {
		rdsClient = aws_clients.NewRDS(awsSession)
	} else {
		fmt.Println("Error while getting aws session")
		os.Exit(0)
	}

	if rdsClient == nil {
		fmt.Println("Error while getting rds client session")
		os.Exit(0)
	}

	dbInstanceInput := &rds.DescribeDBInstancesInput{}
	if len(clusters) > 0 {
		filter := &rds.Filter{}
		filter.Name = aws.String("db-instance-id")
		filter.Values = clusters
		dbInstanceInput.Filters = []*rds.Filter{filter}
	}
	dbClusterOutput, err := rdsClient.DescribeDBInstances(dbInstanceInput)

	if err != nil {
		fmt.Println("An error occurred processing AWS RDS API DescribeDBInstances", err)
		return
	}

	if dbClusterOutput == nil || dbClusterOutput.DBInstances == nil || len(dbClusterOutput.DBInstances) == 0 {
		fmt.Println("UNKNOWN : DB Instance not found!")
		return
	}

	for _, dbInstance := range dbClusterOutput.DBInstances {
		fullScheme := *dbInstance.DBInstanceIdentifier
		if len(scheme) > 0 {
			fullScheme = fmt.Sprintf("%s.%s", scheme, fullScheme)
		}

		cloudWatchClient = aws_clients.NewCloudWatch(awsSession)

		if cloudWatchClient == nil {
			fmt.Println("Failed to create cloud watch client")
			return
		}

		for statistic, _ := range statisticsTypeMap {
			value, timestamp, err := getCloudWatchMetrics(statistic, fullScheme)
			if err != nil {
				fmt.Println("Error : ", err)
				return
			}
			if value == nil || timestamp == nil {
				continue
			}
			fmt.Println(fullScheme, ".", statistic, "  -  value :", value, ", timestamp:", timestamp)
		}
	}
}

func getCloudWatchMetrics(metricname string, rdsName string) (*float64, *time.Time, error) {
	var input cloudwatch.GetMetricStatisticsInput
	input.Namespace = aws.String("AWS/RDS")
	input.MetricName = aws.String(metricname)
	var dimensionFilter cloudwatch.Dimension
	dimensionFilter.Name = aws.String("DBInstanceIdentifier")
	dimensionFilter.Value = aws.String(rdsName)
	input.Dimensions = []*cloudwatch.Dimension{&dimensionFilter}
	input.EndTime = aws.Time(time.Now().Add(time.Duration(-fetchAge/60) * time.Minute))
	input.StartTime = aws.Time((*input.EndTime).Add(time.Duration(-period/60) * time.Minute))
	input.Period = aws.Int64(period)
	input.Statistics = []*string{aws.String(strings.Title(statistics))}
	metrics, err := cloudWatchClient.GetMetricStatistics(&input)
	if err != nil {
		return nil, nil, err
	}
	if metrics != nil && metrics.Datapoints != nil && len(metrics.Datapoints) > 1 {
		var minimumTimeDifference float64
		var timeDifference float64
		var averageValue *float64
		var timestamp *time.Time
		minimumTimeDifference = -1
		for _, datapoint := range metrics.Datapoints {
			timeDifference = time.Since(*datapoint.Timestamp).Seconds()
			if minimumTimeDifference == -1 {
				minimumTimeDifference = timeDifference
				averageValue = datapoint.Average
				timestamp = datapoint.Timestamp
			} else if timeDifference < minimumTimeDifference {
				minimumTimeDifference = timeDifference
				averageValue = datapoint.Average
				timestamp = datapoint.Timestamp
			}
		}
		return averageValue, timestamp, nil
	}
	return nil, nil, nil
}
