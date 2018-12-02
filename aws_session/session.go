package aws_session

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

func CreateAwsSession() *session.Session {

	// Create a Session with a custom region
	aws_session := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-east-2"),
	}))

	return aws_session
}
