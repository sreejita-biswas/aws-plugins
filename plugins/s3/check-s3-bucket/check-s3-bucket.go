package main

/*
#
# check-s3-bucket
#
# DESCRIPTION:
#   This plugin checks a bucket and alerts if not exists
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
#
# USAGE:
#   ./check-s3-bucket --bucket_name=mybucket
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
	"strings"

	"github.com/sreejita-biswas/aws-plugins/awsclient"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
)

var (
	s3Client   *s3.S3
	awsRegion  string
	bucketName string
)

func main() {
	var success bool
	getFlags()
	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)

	success, s3Client = awsclient.GetS3Client(awsSession)
	if !success {
		return
	}
	if len(strings.TrimSpace(bucketName)) == 0 {
		fmt.Println("Enter a bucket name")
		return
	}
	input := &s3.HeadBucketInput{Bucket: aws.String(bucketName)}
	_, err := s3Client.HeadBucket(input)
	if err != nil && err.(awserr.Error).Code() == "NotFound" {
		fmt.Println("CRITICAL:", bucketName, "bucket not found")
	} else if err != nil {
		fmt.Println("CRITICAL:", bucketName, "-", err.(awserr.Error).Message())
	} else {
		fmt.Println("OK:", bucketName, "bucket found")
	}
}

func getFlags() {
	flag.StringVar(&awsRegion, "aws_region", "us-east-1", "AWS Region (defaults to us-east-1).")
	flag.StringVar(&bucketName, "bucket_name", "", "A comma seperated list of S3 buckets to check")
	flag.Parse()
}
