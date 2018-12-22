package main

/*
#
# check-ebs-snapshots
#
# DESCRIPTION:
#   Check EC2 Attached Volumes for Snapshots.  Only for Volumes with a Name tag.
#
# OUTPUT:
#   plain-text
#
# PLATFORMS:
#   MAC OS
#
# USAGE:
#   ./check-ebs-snapshots --check_ignored=false
#
# NOTES:
#   When using check_ignored flag value as true, any volume that has a tag-key of "IGNORE_BACKUP" will
#   be ignored.
#
# LICENSE:
#   TODO
#
*/

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/sreejita-biswas/aws-plugins/aws_clients"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
)

var (
	ec2Client         *ec2.EC2
	scheme            string
	awsRegion         string
	cloudWatchClient  *cloudwatch.CloudWatch
	criticalThreshold float64
	checkIgnored      bool
	period            int64
)

func main() {
	flag.StringVar(&awsRegion, "aws_region", "us-east-2", "AWS Region (defaults to us-east-1).")
	flag.BoolVar(&checkIgnored, "check_ignored", true, "mark as true to ignore volumes with an IGNORE_BACKUP tag")
	flag.Int64Var(&period, "period", 7, "Length in time to alert on missing snapshots")
	flag.Parse()

	var errors []string

	volumeInput := &ec2.DescribeVolumesInput{}
	tagNames := []string{}

	filter := &ec2.Filter{}
	filter.Name = aws.String("attachment.status")
	filter.Values = []*string{aws.String("attached")}
	volumeInput.Filters = []*ec2.Filter{filter}
	filter2 := &ec2.Filter{}
	filter2.Name = aws.String("tag-key")
	filter2.Values = []*string{aws.String("Name")}

	awsSession := aws_session.CreateAwsSessionWithRegion(awsRegion)

	if awsSession != nil {
		ec2Client = aws_clients.NewEC2(awsSession)
	} else {
		fmt.Println("Error while getting aws session")
		os.Exit(0)
	}

	if ec2Client == nil {
		fmt.Println("Error while getting ec2 client session")
		os.Exit(0)
	}

	cloudWatchClient = aws_clients.NewCloudWatch(awsSession)

	if cloudWatchClient == nil {
		fmt.Println("Error while getting cloudwatch client session")
		os.Exit(0)
	}

	volumes, err := ec2Client.DescribeVolumes(volumeInput)
	if err != nil {
		fmt.Println(err)
	}

	if volumes != nil {
		for _, volume := range volumes.Volumes {
			tags := volume.Tags
			ignoreVolume := false
			tagNames = []string{}
			if volume.Tags != nil && len(volume.Tags) > 0 {
				for _, tag := range tags {
					if checkIgnored && *tag.Key == "IGNORE_BACKUP" {
						ignoreVolume = true
						break
					} else {
						tagNames = append(tagNames, *tag.Key)
					}
				}
				if ignoreVolume {
					continue
				}
				latestSnapshot, err := getLatestSnapshot(*volume.VolumeId)
				if err != nil {
					fmt.Println("Error : ", err)
					return
				}
				if latestSnapshot != nil {
					timeDiffrence := aws.Time(time.Now().Add(time.Duration(period*24*60) * time.Minute)).Sub(*latestSnapshot.StartTime)
					if timeDiffrence.Seconds() < 0 {
						errors = append(errors, fmt.Sprintf("%v latest snapshot is %v for Voulme %s \n", tagNames, *latestSnapshot.StartTime, *volume.VolumeId))
					}
				} else {
					errors = append(errors, fmt.Sprintf("%v has no snapshot for Voulme %s \n", tagNames, *volume.VolumeId))
				}
			}
		}

		if len(errors) > 0 {
			fmt.Println("Warning : ", errors)
		} else {
			fmt.Println("Ok")
		}
	}
}

func getLatestSnapshot(volumeId string) (*ec2.Snapshot, error) {
	var latestSnapshot *ec2.Snapshot
	filter := &ec2.Filter{}
	filter.Name = aws.String("volume-id")
	filter.Values = []*string{&volumeId}
	snapshotInput := &ec2.DescribeSnapshotsInput{}
	snapshotInput.Filters = []*ec2.Filter{filter}
	snapshots, err := ec2Client.DescribeSnapshots(snapshotInput)
	if err != nil {
		return nil, err
	}
	if snapshots != nil && snapshots.Snapshots != nil && len(snapshots.Snapshots) >= 1 {
		var minimumTimeDifference float64
		var timeDifference float64
		minimumTimeDifference = -1
		for _, snapshot := range snapshots.Snapshots {
			timeDifference = time.Since(*snapshot.StartTime).Seconds()
			if minimumTimeDifference == -1 {
				minimumTimeDifference = timeDifference
				latestSnapshot = snapshot
			} else if timeDifference < minimumTimeDifference {
				minimumTimeDifference = timeDifference
				latestSnapshot = snapshot
			}
		}

	} else {
		return nil, nil
	}
	return latestSnapshot, err
}
