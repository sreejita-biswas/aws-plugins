package aws_session

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
)

var aws_session *session.Session

func CreateAwsSession() *session.Session {

	// Create a Session with a custom region
	aws_session := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("us-west-2"),
	}))

	// aws_session := session.Must(session.NewSessionWithOptions(session.Options{
	// 	SharedConfigState: session.SharedConfigEnable,
	// }))

	return aws_session
}

func GetAwsSession() *session.Session {
	return aws_session
}
