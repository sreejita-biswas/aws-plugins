package main

/*
#
# check-elb-health-fog
#
#
# DESCRIPTION:
#   This plugin checks the health of an Amazon Elastic Load Balancer.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
#
# USAGE:
#  ./check-elb-health-fog -aws_region=${you_region} --instances=${your_instance_ids} --elb_name=${your_elb_name} --verbose=true
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

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/sreejita-biswas/aws-plugins/aws_clients"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
)

var (
	awsRegion string
	elbName   string
	instances string
	verbose   bool
	elbClient *elb.ELB
)

func main() {
	flag.StringVar(&awsRegion, "aws_region", "eu-west-1", "AWS Region (such as eu-west-1). If you do not specify a region, it will be detected by the server the script is run on")
	flag.StringVar(&elbName, "elb_name", "", "The Elastic Load Balancer name of which you want to check the health")
	flag.StringVar(&instances, "instances", "", "Comma separated list of specific instances IDs inside the ELB of which you want to check the health")
	flag.BoolVar(&verbose, "verbose", false, "Enable a little bit more verbose reports about instance health")
	flag.Parse()

	if len(elbName) <= 0 {
		fmt.Println("Enter a valid load balancer name")
		return
	}

	awsSession := aws_session.CreateAwsSession()

	if awsSession != nil {
		elbClient = aws_clients.NewELB(awsSession)
	} else {
		fmt.Errorf("Error while getting aws session")
		os.Exit(0)
	}

	if elbClient == nil {
		fmt.Errorf("Error while getting elb client session")
		os.Exit(0)
	}

	instanceIdentifiers := strings.Split(instances, ",")

	input := &elb.DescribeInstanceHealthInput{}
	for _, instanceId := range instanceIdentifiers {
		input.Instances = append(input.Instances, &elb.Instance{InstanceId: &instanceId})
	}
	input.LoadBalancerName = &elbName
	output, err := elbClient.DescribeInstanceHealth(input)
	if err != nil {
		fmt.Println("An issue occured while communicating with the AWS EC2 API,", err)
		return
	}
	if !(output != nil && output.InstanceStates != nil && len(output.InstanceStates) > 0) {
		return
	}

	unhealthyInstances := make(map[string]string)
	for _, instanceState := range output.InstanceStates {
		if *instanceState.State != "InService" {
			unhealthyInstances[*instanceState.InstanceId] = *instanceState.State
		}
	}

	if unhealthyInstances == nil || len(unhealthyInstances) <= 0 {
		fmt.Println("OK : All instances on ELB ", awsRegion, "::", elbName, "healthy!")
		return
	}

	if verbose {
		fmt.Println("CRITICAL : Unhealthy instances detected:")
		for id, state := range unhealthyInstances {
			fmt.Println(id, "::", state)
		}
	} else {
		fmt.Println("CRITICAL : Detected ", len(unhealthyInstances), "unhealthy instances")
	}
}
