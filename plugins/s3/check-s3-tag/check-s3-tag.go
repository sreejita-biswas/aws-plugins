package main

/*
#
# check-s3-tag
#
# DESCRIPTION:
#   This plugin checks if buckets have a set of tags.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
# USAGE:
#   ./check-s3-tag --tag_keys=sensu
#
# LICENSE:
#   TODO
#
*/

import (
	"flag"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
	"github.com/sreejita-biswas/aws-plugins/awsclient"
)

var (
	s3Client  *s3.S3
	tagKeys   string
	awsRegion string
)

func main() {
	var success bool
	flag.StringVar(&awsRegion, "aws_region", "us-east-1", "AWS Region (defaults to us-east-1).")
	flag.StringVar(&tagKeys, "tag_keys", "", "Comma seperated Tag Keys")
	flag.Parse()

	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)
	success, s3Client = awsclient.GetS3Client(awsSession)
	if !success {
		return
	}

	if len(strings.TrimSpace(tagKeys)) == 0 {
		fmt.Println("Enter atleast one tag key")
		return
	}

	if len(strings.TrimSpace(tagKeys)) == 1 && strings.TrimSpace(tagKeys) == "," {
		fmt.Println("Enter atleast one tag key")
		return
	}

	tags := strings.Split(tagKeys, ",")
	tagMap := make(map[string]*string)
	missingTagsMap := make(map[string][]string)

	for _, tag := range tags {
		tagMap[tag] = &tag
	}

	input := &s3.ListBucketsInput{}
	output, err := s3Client.ListBuckets(input)
	if err != nil {
		fmt.Println(err)
		return
	}
	if output != nil && output.Buckets != nil && len(output.Buckets) > 1 {
		for _, bucket := range output.Buckets {
			bucketInput := &s3.GetBucketTaggingInput{Bucket: bucket.Name}
			bucketOutput, err := s3Client.GetBucketTagging(bucketInput)
			if err != nil {
				continue
			}
			if bucketOutput != nil && bucketOutput.TagSet != nil && len(bucketOutput.TagSet) > 0 {
				bucketTagMap := make(map[string]*string)
				for _, bucketTag := range bucketOutput.TagSet {
					bucketTagMap[*bucketTag.Key] = bucketTag.Key
				}
				for tag, tagValue := range tagMap {
					if bucketTagMap[tag] != nil {
						continue
					} else {
						missingTagsMap[*bucket.Name] = append(missingTagsMap[*bucket.Name], *tagValue)
					}
				}
			}
		}
	}

	if len(missingTagsMap) > 0 {
		for bucketName, tags := range missingTagsMap {
			fmt.Println("CRITICAL : Missing tags for bucket", bucketName, ":", tags)
		}
	}
}
