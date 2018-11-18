package plugins

import (
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/sreejita-biswas/aws-plugins/aws_clients"
	"time"
)

type IAMServerCertificate struct{
	Name string
	Expiration time.Time
}

func ListServerCertificates(certificateName *string) ([]IAMServerCertificate, error) {
	listServerCertificatesInput := &iam.ListServerCertificatesInput{}
	iamClient := aws_clients.GetIAMAwsClient()

	listServerCertificatesOutput, err := iamClient.ListServerCertificates(listServerCertificatesInput)
	if err != nil {
		return []IAMServerCertificate{}, err
	}

	certificates := make([]IAMServerCertificate, 0, len(listServerCertificatesOutput.ServerCertificateMetadataList))
	for _, metadata := range listServerCertificatesOutput.ServerCertificateMetadataList {
		certificate := IAMServerCertificate{
			Name:       *metadata.ServerCertificateName,
			Expiration: *metadata.Expiration,
		}
		if certificateName != nil && *certificateName == *metadata.ServerCertificateName {

			return []IAMServerCertificate{certificate},nil
		}
		certificates = append(certificates, certificate)
	}

	return certificates, nil
}

