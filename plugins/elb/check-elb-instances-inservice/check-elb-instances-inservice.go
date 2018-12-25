package main

/*
#
# check-elb-instance-inservice
#
# DESCRIPTION:
#   Check Elastic Loudbalancer Instances are inService.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
#
# USAGE:
#   all LoadBalancers
#   ./check-elb-instance-inservice --aws_region=${your_region}
#   one loadBalancer
#   ./check-elb-instance-inservice --aws_region=${your_region} --elb_name=${LoadBalancerName}
#
# NOTES:
#   Based heavily on Peter Hoppe check-autoscaling-instances-inservices
#
# LICENSE:
#  TODO
#
*/

import (
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/sreejita-biswas/aws-plugins/aws_clients"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
)

var (
	awsRegion string
	elbName   string
	elbClient *elb.ELB
	ec2Client *ec2.EC2
)

func main() {
	flag.StringVar(&awsRegion, "aws_region", "eu-west-1", "AWS Region (such as eu-west-1). If you do not specify a region, it will be detected by the server the script is run on")
	flag.StringVar(&elbName, "elb_name", "", "The Elastic Load Balancer name of which you want to check the health")
	flag.Parse()

	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)

	if awsSession != nil {
		elbClient = aws_clients.NewELB(awsSession)
	} else {
		fmt.Errorf("Error while getting aws session")
		os.Exit(0)
	}

	ec2Client = aws_clients.NewEC2(awsSession)

	if ec2Client == nil {
		fmt.Println("Error while getting ec2 client session")
		os.Exit(0)
	}

	if elbClient == nil {
		fmt.Errorf("Error while getting elb client session")
		os.Exit(0)
	}

	//Find all load balancers/specific load balance specific to the given region
	input := &elb.DescribeLoadBalancersInput{}
	if len(elbName) > 0 {
		input.LoadBalancerNames = []*string{&elbName}
	}
	output, err := elbClient.DescribeLoadBalancers(input)
	if err != nil {
		fmt.Println("An issue occured while communicating with the AWS EC2 API,", err)
		return
	}

	if !(output != nil && output.LoadBalancerDescriptions != nil && len(output.LoadBalancerDescriptions) > 0) {
		fmt.Println("No Load Balancer found in region -", awsRegion)
		return
	}

	for _, loadBalancer := range output.LoadBalancerDescriptions {
		unhealthyInstances := make(map[string]string)
		healtStatusInput := &elb.DescribeInstanceHealthInput{}
		healtStatusInput.LoadBalancerName = loadBalancer.LoadBalancerName
		healtStatusOutput, err := elbClient.DescribeInstanceHealth(healtStatusInput)
		if err != nil {
			fmt.Println("An issue occured while communicating with the AWS EC2 API,", err)
			return
		}
		if !(output != nil && healtStatusOutput.InstanceStates != nil && len(healtStatusOutput.InstanceStates) > 0) {
			continue
		}

		for _, instanceState := range healtStatusOutput.InstanceStates {
			if *instanceState.State != "InService" {
				unhealthyInstances[*instanceState.InstanceId] = *instanceState.State
			}
		}
		if len(unhealthyInstances) == 0 {
			fmt.Println("OK : All instances of Load Balancer - ", *loadBalancer.LoadBalancerName, "are in healthy state")
			continue
		}

		if len(unhealthyInstances) == len(healtStatusOutput.InstanceStates) {
			fmt.Println("CRITICAL : All instances of Load Balancer - ", *loadBalancer.LoadBalancerName, "are in unhealthy state")
			continue
		}

		fmt.Println("WARNING : Unhealthy Instances for Load Balanacer - ", *loadBalancer.LoadBalancerName, "are")
		for id, state := range unhealthyInstances {
			fmt.Println("Instace - ", id, ":: State - ", state)
		}

	}
}
