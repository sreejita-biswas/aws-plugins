package config

import "github.com/micro/cli"

type Config struct {
	AwsAcessKey         string `key:"aws_access_key"`
	AwsSecretAccessKey  string `key:"aws_secret_access_key"`
	AwsRegion           string `key:"aws_region"`
	ServerCerticateName string `key:"server_certificate_name"`
}

func Flags(conf *Config) []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "aws_access_key",
			Value:       "",
			Usage:       "AWS Access Key. Either set ENV['AWS_ACCESS_KEY'] or provide it as an option. Uses Default Credential if none are passed",
			EnvVar:      "AWS_ACCESS_KEY",
			Destination: &conf.AwsAcessKey,
		},
		cli.StringFlag{
			Name:        "aws_secret_access_key",
			Value:       "",
			Usage:       "AWS Secret Access Key. Either set ENV['AWS_SECRET_KEY'] or provide it as an option. Uses Default Credential if none are passed",
			EnvVar:      "AWS_SECRET_KEY",
			Destination: &conf.AwsSecretAccessKey,
		},
		cli.StringFlag{
			Name:        "aws_region",
			Value:       "us-east-1",
			Usage:       "AWS Region (defaults to us-east-1).",
			EnvVar:      "AWS_REGION",
			Destination: &conf.AwsRegion,
		},
		cli.StringFlag{
			Name:        "server_certificate_name",
			Value:       "",
			Usage:       "Certificate to check. Checks all if not passed",
			EnvVar:      "CERTIFICATE_NAME",
			Destination: &conf.ServerCerticateName,
		},
	}
}
