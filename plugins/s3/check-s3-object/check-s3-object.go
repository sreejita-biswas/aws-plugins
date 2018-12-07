package main

/*
#
# check-s3-object
#
# DESCRIPTION:
#   This plugin checks if a file exists in a bucket and/or is not too old.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
#
# USAGE:
#   ./check-s3-object.rb --bucket-name mybucket --aws-region eu-west-1 --use-iam --key-name "path/to/myfile.txt"
#   ./check-s3-object.rb --bucket-name mybucket --aws-region eu-west-1 --use-iam --key-name "path/to/myfile.txt" --warning 90000 --critical 126000
#   ./check-s3-object.rb --bucket-name mybucket --aws-region eu-west-1 --use-iam --key-name "path/to/myfile.txt" --warning 90000 --critical 126000 --ok-zero-size
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
	"time"

	"github.com/sreejita-biswas/aws-plugins/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"

	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/sreejita-biswas/aws-plugins/aws_clients"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
)

var (
	s3Client                *s3.S3
	filters                 string
	useIamRole              bool
	bucketName              string
	allBuckets              bool
	excludeBuckets          string
	keyName                 string
	keyPrefix               string
	warningAge              float64
	criticalAge             float64
	okZeroSize              bool
	warningSize             int64
	criticalSize            int64
	compareSize             string
	noCritOnMultipleObjects bool
	awsRegion               string
)

func main() {
	flag.StringVar(&awsRegion, "aws_region", "us-east-2", "AWS Region (defaults to us-east-1).")
	flag.BoolVar(&useIamRole, "use_iam_role", false, "Use IAM role authenticiation. Instance must have IAM role assigned for this to work")
	flag.StringVar(&bucketName, "bucket_name", "sreejita-testing", "The name of the S3 bucket where object lives")
	flag.StringVar(&keyName, "key_name", "", "The name of key in the bucket")
	flag.StringVar(&keyPrefix, "key_prefix", "s3", "Prefix key to search on the bucket")
	flag.Float64Var(&warningAge, "warning_age", 90000, "Warn if mtime greater than provided age in seconds")
	flag.Float64Var(&criticalAge, "critical_age", 126000, "Critical if mtime greater than provided age in seconds")
	flag.BoolVar(&okZeroSize, "ok_zero_size", true, "OK if file has zero size'")
	flag.Int64Var(&warningSize, "warning_size", 0, "Warning threshold for size")
	flag.Int64Var(&criticalSize, "critical_size", 0, "Critical threshold for size")
	flag.StringVar(&compareSize, "operator-size", "equal", "Comparision operator for threshold: equal, not, greater, less")
	flag.BoolVar(&noCritOnMultipleObjects, "no_crit_on_multiple_objects", true, "If this flag is set, sort all matching objects by last_modified date and check against the newest. By default, this check will return a CRITICAL result if multiple matching objects are found.")
	flag.Parse()

	var age time.Duration
	var size int64
	var keyFullName string

	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)

	if awsSession != nil {
		s3Client = aws_clients.NewS3(awsSession)
	} else {
		fmt.Println("Error while getting aws session")
		os.Exit(0)
	}

	if s3Client == nil {
		fmt.Println("Error while getting s3 client session")
		os.Exit(0)
	}

	if len(strings.TrimSpace(bucketName)) == 0 {
		fmt.Println("Enter a bucket name")
	}

	if (len(strings.TrimSpace(keyName)) == 0 && len(strings.TrimSpace(keyPrefix)) == 0) || (len(strings.TrimSpace(keyName)) > 0 && len(strings.TrimSpace(keyPrefix)) > 0) {
		fmt.Println("Need one option between \"key_name\" and \"key_prefix\"")
	}

	if len(strings.TrimSpace(keyName)) > 0 {
		input := &s3.HeadObjectInput{Bucket: aws.String(bucketName), Key: aws.String(keyName)}
		output, err := s3Client.HeadObject(input)
		if err != nil {
			printErroMessage(err, keyFullName)
			return
		}
		if output != nil {
			age = time.Since(*output.LastModified)
			size = *output.ContentLength
			keyFullName = keyName
			printMesaage(age, keyFullName, size)
		}
	} else if len(strings.TrimSpace(keyPrefix)) > 0 {
		input := &s3.ListObjectsInput{Bucket: aws.String(bucketName), Prefix: aws.String(keyPrefix)}
		output, err := s3Client.ListObjects(input)
		if err != nil {
			printErroMessage(err, keyFullName)
			return
		}
		if output == nil || output.Contents == nil || len(output.Contents) < 1 {
			fmt.Println("CRITICAL : Object with prefix ", keyPrefix, "not found in bucket ", bucketName)
			return
		}

		if output != nil || output.Contents != nil || len(output.Contents) > 1 {
			if !noCritOnMultipleObjects {
				fmt.Println("CRITICAL : Your prefix \"", keyPrefix, "\" return too much files, you need to be more specific")
				return
			} else {
				utils.SortContents(output.Contents)
			}
		}

		keyFullName = *output.Contents[0].Key
		age = time.Since(*output.Contents[0].LastModified)
		size = *output.Contents[0].Size
		printMesaage(age, keyFullName, size)
	}
}

func checkAge(age time.Duration, keyName string) {
	if age.Seconds() > criticalAge {
		fmt.Println("CRITICAL : S3 Object", keyName, "is", age.Seconds(), "seconds old (Bucket -", bucketName, ")")
		return
	}
	if age.Seconds() > warningAge {
		fmt.Println("WARNING : S3 Object", keyName, "is", age.Seconds(), " seconds old (Bucket -", bucketName, ")")
		return
	}
	fmt.Println("OK : S3 Object", keyName, "exists in bucket", bucketName)

}

func checkSize(size int64, keyName string) {
	if compareSize == "not" {
		if size != criticalSize {
			fmt.Println("CRITICAL : S3 Object", keyName, "size :", size, "octets (Bucket - ", bucketName, ")")
			return
		}
		if size != warningSize {
			fmt.Println("WARNING : S3 Object", keyName, "size :", size, "octets (Bucket - ", bucketName, ")")
			return
		}
	}

	if compareSize == "greater" {
		if size > criticalSize {
			fmt.Println("CRITICAL : S3 Object", keyName, "size :", size, "octets (Bucket - ", bucketName, ")")
			return
		}
		if size > warningSize {
			fmt.Println("WARNING : S3 Object", keyName, "size :", size, "octets (Bucket - ", bucketName, ")")
			return
		}

		fmt.Println("OK : S3 Object", keyName, "exists in bucket", bucketName)
	}

	if compareSize == "less" {
		if size < criticalSize {
			fmt.Println("CRITICAL : S3 Object", keyName, "size :", size, "octets (Bucket -", bucketName, ")")
			return
		}
		if size < warningSize {
			fmt.Println("WARNING : S3 Object", keyName, "size :", size, "octets (Bucket -", bucketName, ")")
			return
		}
		fmt.Println("OK : S3 Object", keyName, "exists in bucket", bucketName)
	}

}

func printErroMessage(err error, keyFullName string) {
	if err.(awserr.Error).Code() == "NotFound" {
		fmt.Println("CRITICAL : S3 Object", keyFullName, "not found in bucket -", bucketName)
	} else {
		fmt.Println("CRITICAL : S3 Object", keyFullName, "in bucket -", bucketName, ",", err.(awserr.Error).Code(), "-", err.(awserr.Error).Message())
	}
}

func printMesaage(age time.Duration, keyFullName string, size int64) {
	checkAge(age, keyFullName)
	if size != 0 {
		checkSize(size, keyFullName)
	} else if !okZeroSize {
		fmt.Println("CRITICAL : S3 Object", keyFullName, "is empty (Bucket -", bucketName, ")")
	}
}
