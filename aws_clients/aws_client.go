package aws_clients

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
)

func NewIAM(awsSession *session.Session) *iam.IAM {
	iamClient := iam.New(awsSession)
	return iamClient
}

func NewEC2(awsSession *session.Session) *ec2.EC2 {
	ec2Client := ec2.New(awsSession)
	return ec2Client
}

func NewCloudWatch(awsSession *session.Session) *cloudwatch.CloudWatch {
	cloudwatchClient := cloudwatch.New(awsSession)
	return cloudwatchClient
}
