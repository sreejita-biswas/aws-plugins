package awsclient

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/aws/aws-sdk-go/service/s3"
)

func newIAM(awsSession *session.Session) *iam.IAM {
	iamClient := iam.New(awsSession)
	return iamClient
}

func newEC2(awsSession *session.Session) *ec2.EC2 {
	ec2Client := ec2.New(awsSession)
	return ec2Client
}

func newCloudWatch(awsSession *session.Session) *cloudwatch.CloudWatch {
	cloudwatchClient := cloudwatch.New(awsSession)
	return cloudwatchClient
}

func newS3(awsSession *session.Session) *s3.S3 {
	s3Client := s3.New(awsSession)
	return s3Client
}

func newRDS(awsSession *session.Session) *rds.RDS {
	rdsClient := rds.New(awsSession)
	return rdsClient
}

func newELB(awsSession *session.Session) *elb.ELB {
	elbClient := elb.New(awsSession)
	return elbClient
}

func newELBV2(awsSession *session.Session) *elbv2.ELBV2 {
	elbv2Client := elbv2.New(awsSession)
	return elbv2Client
}

func GetElbClient(awsSession *session.Session) (bool, *elb.ELB) {
	var elbClient *elb.ELB
	if awsSession != nil {
		elbClient = newELB(awsSession)
	} else {
		fmt.Println("Error while getting aws session")
		return false, nil
	}

	if elbClient == nil {
		fmt.Println("Error while getting elb client session")
		return false, nil
	}

	return true, elbClient
}

func GetElbV2Client(awsSession *session.Session) (bool, *elbv2.ELBV2) {
	var elbClient *elbv2.ELBV2
	if awsSession != nil {
		elbClient = newELBV2(awsSession)
	} else {
		fmt.Println("Error while getting aws session")
		return false, nil
	}

	if elbClient == nil {
		fmt.Println("Error while getting elbv2 client session")
		return false, nil
	}

	return true, elbClient
}

func GetEC2Client(awsSession *session.Session) (bool, *ec2.EC2) {
	var ec2Client *ec2.EC2
	if awsSession != nil {
		ec2Client = newEC2(awsSession)
	} else {
		fmt.Println("Error while getting aws session")
		return false, nil
	}

	if ec2Client == nil {
		fmt.Println("Error while getting ec2 client session")
		return false, nil
	}

	return true, ec2Client
}

func GetCloudWatchClient(awsSession *session.Session) (bool, *cloudwatch.CloudWatch) {
	var cloudWatClient *cloudwatch.CloudWatch
	if awsSession != nil {
		cloudWatClient = newCloudWatch(awsSession)
	} else {
		fmt.Println("Error while getting aws session")
		return false, nil
	}

	if cloudWatClient == nil {
		fmt.Println("Error while getting cloudwatch client session")
		return false, nil
	}

	return true, cloudWatClient
}

func GetRDSClient(awsSession *session.Session) (bool, *rds.RDS) {
	var rdsClient *rds.RDS
	if awsSession != nil {
		rdsClient = newRDS(awsSession)
	} else {
		fmt.Println("Error while getting aws session")
		return false, nil
	}

	if rdsClient == nil {
		fmt.Println("Error while getting rds client session")
		return false, nil
	}

	return true, rdsClient
}

func GetS3Client(awsSession *session.Session) (bool, *s3.S3) {
	var s3Client *s3.S3
	if awsSession != nil {
		s3Client = newS3(awsSession)
	} else {
		fmt.Println("Error while getting aws session")
		return false, nil
	}

	if s3Client == nil {
		fmt.Println("Error while getting s3 client session")
		return false, nil
	}

	return true, s3Client
}
