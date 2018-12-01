package aws_clients

import (
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
)

var iam_client *iam.IAM
var ec2_client *ec2.EC2
var cloudwatch_client *cloudwatch.CloudWatch

func NewIAM() *iam.IAM {
	iam_client = iam.New(aws_session.GetAwsSession())
	return iam_client
}

func NewEC2(awsSession *session.Session) *ec2.EC2 {
	ec2_client = ec2.New(awsSession)
	return ec2_client
}

func NewCloudWatch() *cloudwatch.CloudWatch {
	cloudwatch_client = cloudwatch.New(aws_session.GetAwsSession())
	return cloudwatch_client
}

func GetIAMAwsClient() *iam.IAM {
	return iam_client
}

func GetEC2AwsClient() *ec2.EC2 {
	return ec2_client
}

func GetCloudwatchAwsClient() *cloudwatch.CloudWatch {
	return cloudwatch_client
}
