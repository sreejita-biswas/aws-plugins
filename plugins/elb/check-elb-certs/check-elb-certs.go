package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"strings"

	"github.com/sreejita-biswas/aws-plugins/awsclient"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
)

/*
#
# check-elb-certs
#
# DESCRIPTION:
#   This plugin looks up all ELBs in the region and checks https
#   endpoints for expiring certificates
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
#
# USAGE:
#  ./check-elb-certs -aws_region=${your_region} -warning=${days_to_warn} -critical=${days_to_critical}
#
# NOTES:
#
# LICENSE:
#   TODO
#
*/

var (
	awsRegion string
	warning   int
	critical  int
	verbose   bool
	elbClient *elb.ELB
)

func main() {
	getFlags()

	awsSession := aws_session.CreateAwsSession()

	success, elbClient := awsclient.GetElbClient(awsSession)
	if !success {
		return
	}

	describeLoadBalancerInput := &elb.DescribeLoadBalancersInput{}
	describeLoadBalancerOutput, err := elbClient.DescribeLoadBalancers(describeLoadBalancerInput)
	if err != nil {
		fmt.Println("Error :", err)
		return
	}
	if !(describeLoadBalancerOutput != nil && describeLoadBalancerOutput.LoadBalancerDescriptions != nil && len(describeLoadBalancerOutput.LoadBalancerDescriptions) > 0) {
		return
	}
	for _, loadBalancer := range describeLoadBalancerOutput.LoadBalancerDescriptions {
		for _, listener := range loadBalancer.ListenerDescriptions {
			elbListener := listener.Listener
			if strings.ToUpper(*elbListener.Protocol) == "HTTPS" {
				dnsName := *loadBalancer.DNSName
				fmt.Println(dnsName)
				port := *elbListener.LoadBalancerPort
				ips, err := net.LookupIP(dnsName)
				if err != nil {
					fmt.Println("Error :", err)
					return
				}
				dialer := net.Dialer{}
				connection, err := tls.DialWithDialer(&dialer, "tcp", fmt.Sprintf("[%s]:%s", ips[0], port), &tls.Config{ServerName: dnsName})
				if err != nil {
					fmt.Println("Error :", err)
					return
				}
				for _, chain := range connection.ConnectionState().VerifiedChains {
					for _, cert := range chain {
						if cert.IsCA {
							continue
						}
					}
				}
			}
		}
	}
}
func getFlags() {
	flag.StringVar(&awsRegion, "aws_region", "us-west-1", "AWS Region (defaults to us-east-1).")
	flag.IntVar(&warning, "warning", 30, "Warn on minimum number of days to SSL/TLS certificate expiration")
	flag.IntVar(&critical, "critical", 5, "Minimum number of days to SSL/TLS certificate expiration")
	flag.BoolVar(&verbose, "verbose", false, "Provide SSL/TLS certificate expiration details even when OK")
	flag.Parse()
}
