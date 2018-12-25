package main

/*
# check-elb-health-sdk
#
# DESCRIPTION:
#   This plugin checks the health of an Amazon Elastic Load Balancer in a given region.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
# USAGE:
#   ./check-elb-health-sdk --aws_region=region
#   ./check-elb-health-sdk --aws_region=region --elb_name=my-elb
#   ./check-elb-health-sdk --aws_region=region --elb_name=my-elb --instances=instance1,instance2
#
# LICENSE
#  TODO
#
*/

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/sreejita-biswas/aws-plugins/aws_clients"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
)

var (
	awsRegion   string
	elbName     string
	instances   string
	verbose     bool
	instanceTag string
	warnOnly    bool
	elbClient   *elb.ELB
	ec2Client   *ec2.EC2
)

func main() {
	flag.StringVar(&awsRegion, "aws_region", "eu-west-1", "AWS Region (such as eu-west-1). If you do not specify a region, it will be detected by the server the script is run on")
	flag.StringVar(&elbName, "elb_name", "", "The Elastic Load Balancer name of which you want to check the health")
	flag.StringVar(&instances, "instances", "", "Comma separated list of specific instances IDs inside the ELB of which you want to check the health")
	flag.BoolVar(&verbose, "verbose", false, "Enable a little bit more verbose reports about instance health")
	flag.StringVar(&instanceTag, "instance_tag", "Name", "Specify instance tag to be included in the check output. E.g. 'Name' tag")
	flag.BoolVar(&warnOnly, "warn_only", false, "Warn instead of critical when unhealthy instances are found")
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

	//Find all load balancers specific to the given region
	input := &elb.DescribeLoadBalancersInput{}
	output, err := elbClient.DescribeLoadBalancers(input)
	if err != nil {
		fmt.Println("An issue occured while communicating with the AWS EC2 API,", err)
		return
	}

	if !(output != nil && output.LoadBalancerDescriptions != nil && len(output.LoadBalancerDescriptions) > 0) {
		fmt.Println("No Load Balancer found in region -", awsRegion)
		return
	}

	elbs := []string{}
	inlcudeElb := false
	allElbs := []string{}
	for _, loadbalancer := range output.LoadBalancerDescriptions {
		if len(elbName) > 0 && *loadbalancer.LoadBalancerName == elbName {
			elbs = append(elbs, elbName)
			inlcudeElb = true
		} else {
			allElbs = append(allElbs, *loadbalancer.LoadBalancerName)
		}
	}

	if !inlcudeElb {
		elbs = allElbs
	}

	critical := false
	elbStateMapping := make(map[string]map[string]string)
	for _, loadBalancer := range elbs {
		unhealthyInstances := make(map[string]string)
		instanceIdentifiers := strings.Split(instances, ",")
		healtStatusInput := &elb.DescribeInstanceHealthInput{}
		for _, instanceId := range instanceIdentifiers {
			healtStatusInput.Instances = append(healtStatusInput.Instances, &elb.Instance{InstanceId: &instanceId})
		}
		healtStatusInput.LoadBalancerName = &loadBalancer
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

				tagInput := &ec2.DescribeTagsInput{}
				filter := &ec2.Filter{}
				filter.Name = aws.String("resource-id")
				filter.Values = []*string{instanceState.InstanceId}
				tagInput.Filters = []*ec2.Filter{filter}
				tagOutput, err := ec2Client.DescribeTags(tagInput)
				if err != nil {
					fmt.Println("An issue occured while communicating with the AWS EC2 API,", err)
					return
				}

				if tagOutput != nil && tagOutput.Tags != nil && len(tagOutput.Tags) > 0 {
					for _, tag := range tagOutput.Tags {
						if *tag.Key == instanceTag {
							unhealthyInstances[*instanceState.InstanceId] = fmt.Sprintf("%s::%s", *tag.Value, *instanceState.State)
							break
						}
					}
				}
			}
		}
		if unhealthyInstances != nil && len(unhealthyInstances) > 0 {
			critical = true
			elbStateMapping[loadBalancer] = unhealthyInstances
			continue
		}
	}

	if !critical && verbose {
		fmt.Println("OK : All instances on all ELBs are healthy!")
		return
	}

	if verbose {
		if warnOnly {
			fmt.Println("WARNING : Unhealthy instances detected: For Elbs")
		} else {
			fmt.Println("CRITICAL : Unhealthy instances detected: For Elbs")
		}
		for loadBlancer, instanceStates := range elbStateMapping {
			fmt.Println("ELB : ", loadBlancer)
			for instaceId, instanceState := range instanceStates {
				fmt.Println(instaceId, "::", instanceState)
			}
		}
	} else {
		if warnOnly {
			fmt.Println("WARNING : Detected ", len(elbStateMapping), "unhealthy elbs")
		} else {
			fmt.Println("CRITICAL : Detected ", len(elbStateMapping), "unhealthy elbs")
		}
	}
}
