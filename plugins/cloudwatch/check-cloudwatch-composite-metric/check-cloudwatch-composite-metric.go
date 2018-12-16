package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/sreejita-biswas/aws-plugins/aws_clients"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
)

/*
#
# check-cloudwatch-composite-metric
#
# DESCRIPTION:
#   This plugin retrieves a couple of values of two cloudwatch metrics,
#   computes a percentage value based on the numerator metric and the denomicator metric
#   and triggers alarms based on the thresholds specified.
#   This plugin is an extension to the Andrew Matheny's check-cloudwatch-metric plugin
#   and uses the CloudwatchCommon lib, extended as well.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
# USAGE:
#   ./check-cloudwatch-composite-metric --namespace AWS/ELB --dimensions="LoadBalancerName=test-elb" --period=60 --statistics=Maximum --operator=equal --critical=0
#
# NOTES:
#
# LICENSE:
#   TODO
#
*/

var (
	excludeAlarms         string
	state                 string
	cloudWatchClient      *cloudwatch.CloudWatch
	awsRegion             string
	namespace             string
	numeratorMetric       bool
	denominatorMetric     bool
	dimensions            string
	period                int64
	statistics            string
	unit                  string
	critical              float64
	warning               float64
	compare               string
	numeratorDefault      float64
	noDenominatorDataOk   bool
	zeroDenominatorDataOk bool
	noDataOk              bool
	numeratorMetricName   string
	denominatorMetricName string
)

func main() {
	flag.StringVar(&excludeAlarms, "exclude_alarms", "", "Exclude alarms")
	flag.StringVar(&state, "state", "ALARM", "State of the alarm")
	flag.StringVar(&awsRegion, "aws_region", "us-east-2", "AWS Region (defaults to us-east-1).")
	flag.StringVar(&namespace, "namespace", "AWS/EC2", "CloudWatch namespace for metric")
	flag.BoolVar(&numeratorMetric, "numerator_metric", true, "Numerator metric name present")
	flag.StringVar(&numeratorMetricName, "numerator_metric_name", "", "Numerator metric name")
	flag.BoolVar(&denominatorMetric, "denominator_metric", true, "Denominator metric name present")
	flag.StringVar(&denominatorMetricName, "denominator_metric_name", "", "Denominator metric name")
	flag.StringVar(&dimensions, "dimensions", "", "Comma delimited list of DimName=Value")
	flag.Int64Var(&period, "period", 60, "CloudWatch metric statistics period in seconds. Must be a multiple of 60")
	flag.StringVar(&statistics, "statistics", "Average", "CloudWatch statistics method")
	flag.StringVar(&unit, "unit", "", "CloudWatch metric unit")
	flag.Float64Var(&critical, "critical", 0, "Trigger a critical when value is over VALUE as a Percent")
	flag.Float64Var(&warning, "warning", 0, "Trigger a warning when value is over VALUE as a Percent")
	flag.StringVar(&compare, "compare", "greater", "Comparision operator for threshold: equal, not, greater, less")
	flag.Float64Var(&numeratorDefault, "numerator_default", 0, "Default for numerator if no data is returned for metric")
	flag.BoolVar(&noDenominatorDataOk, "no_denominator_data_ok", false, "Returns ok if no data is returned from denominator metric")
	flag.BoolVar(&zeroDenominatorDataOk, "zero_denominator_data_ok", false, "Returns ok if denominator metric is zero")
	flag.BoolVar(&noDataOk, "no_data_ok", false, "Returns ok if no data is returned from either metric")
	flag.Parse()

	var numeratorMetricValue *float64
	var denomatorMetricValue *float64
	var err error

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

	if numeratorMetric && len(numeratorMetricName) <= 0 {
		fmt.Println("Provide a valid numerator metric name")
		return
	}

	if denominatorMetric && len(denominatorMetricName) <= 0 {
		fmt.Println("Provide a valid denomitator metric name")
		return
	}

	if len(dimensions) <= 0 || dimensions == "," {
		fmt.Println("Please provide valid dimensions")
	}

	if numeratorMetric {
		numeratorMetricValue, err = geMetrics(numeratorMetricName)
		if err != nil {
			fmt.Println("Error while getting metric statistics for Metric :", numeratorMetricName)
			return
		}
	}

	if denominatorMetric {
		denomatorMetricValue, err = geMetrics(denominatorMetricName)
		if err != nil {
			fmt.Println("Error while getting metric statistics for Metric :", denominatorMetricName)
			return
		}
	}

	if numeratorMetricValue == nil {
		numeratorMetricValue = &numeratorDefault
	}

	// If the numerator is empty, then we see if there is a default. If there is a default
	// then we will pretend the numerator _isnt_ empty. That is
	// if empty but there is no default this will be true. If it is empty and there is a default
	// this will be false (i.e. there is data, following standard of dealing in the negative here)
	noData := (numeratorMetricValue == nil) || (denomatorMetricValue == nil)

	// no data in numerator or denominator this is to keep backwards compatibility
	if noData && noDataOk {
		fmt.Println("OK : Returned no data but that's ok")
		return
	} else if denomatorMetricValue == nil && noDenominatorDataOk {
		fmt.Println("OK : ", denominatorMetricName, "returned no data but that's ok")
		return
	} else if noData { // This is legacy case
		fmt.Println("Unknown : metric data could not be retrieved")
		return
	}

	// Now both the denominator and numerator have data (or a valid default)
	if *denomatorMetricValue == 0 && zeroDenominatorDataOk {
		fmt.Println("OK :", denominatorMetricName, ": denominator value is zero but that's ok")
		return
	} else if *denomatorMetricValue != 0 {
		fmt.Println("Unknown :", denominatorMetricName, ": denominator value is zero")
		return
	}

	// We already checked if this value is nil so we know its not
	value := *numeratorMetricValue / (*denomatorMetricValue) * 100
	message := fmt.Sprintf("%s-%s/%s-(%s) is value %f", namespace, numeratorMetricName, denominatorMetricName, dimensions, value)

	if compare == "greater" {
		if value > critical {
			fmt.Println("CRITICAL : ", message)
			return
		}
		if value > warning {
			fmt.Println("WARNING : ", message)
			return
		}
		fmt.Println("OK : ", message)
		return
	} else if compare == "less" {
		if value < critical {
			fmt.Println("CRITICAL : ", message)
			return
		}
		if value < warning {
			fmt.Println("WARNING : ", message)
			return
		}
		fmt.Println("OK : ", message)
		return
	} else if compare == "equal" {
		if value == critical {
			fmt.Println("CRITICAL : ", message)
			return
		}
		if value == warning {
			fmt.Println("WARNING : ", message)
			return
		}
		fmt.Println("OK : ", message)
		return
	} else if compare == "not" {
		if value != critical {
			fmt.Println("CRITICAL : ", message)
			return
		}
		if value != warning {
			fmt.Println("WARNING : ", message)
			return
		}
		fmt.Println("OK : ", message)
		return
	}
}

func geMetrics(metricName string) (*float64, error) {
	input := &cloudwatch.GetMetricStatisticsInput{}
	input.Namespace = aws.String(namespace)
	input.MetricName = aws.String(metricName)
	dimensionInputs := strings.Split(dimensions, ",")
	for _, dimension := range dimensionInputs {
		dimensionNameValuePair := strings.Split(dimension, "=")
		if dimensionNameValuePair != nil && len(dimensionNameValuePair) == 2 {
			input.Dimensions = append(input.Dimensions, &cloudwatch.Dimension{Name: aws.String(dimensionNameValuePair[0]), Value: aws.String(dimensionNameValuePair[1])})
		}
	}
	input.EndTime = aws.Time(time.Now())
	input.StartTime = aws.Time(time.Now().Add(time.Duration(-10*(period/60)) * time.Minute))
	input.Period = aws.Int64(period)
	input.Statistics = []*string{aws.String(statistics)}
	input.Unit = aws.String(unit)
	var minimumTimeDifference float64
	var timeDifference float64
	var averageValue *float64
	minimumTimeDifference = -1
	metrics, err := cloudWatchClient.GetMetricStatistics(input)
	if err != nil {
		return nil, err
	}
	if metrics != nil && metrics.Datapoints != nil && len(metrics.Datapoints) > 1 {
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
	}
	return averageValue, nil
}
