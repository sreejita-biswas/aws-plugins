package aws_clients

import (
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
)

var iam_client *iam.IAM



func NewIAM() *iam.IAM {
	iam_client =  iam.New(aws_session.GetAwsSession())
	return iam_client
}


func GetIAMAwsClient () (*iam.IAM){
	return iam_client
}