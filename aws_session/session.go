package aws_session

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/sreejita-biswas/aws-plugins/config"
)

var aws_session *session.Session

func CreateAwsSession(config config.Config) *session.Session {

	// Create a Session with a custom region
	aws_session := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(config.AwsRegion),
	}))

	return aws_session
}

func GetAwsSession() *session.Session {
	return aws_session
}
