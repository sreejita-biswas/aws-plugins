package config

import "github.com/micro/cli"

type Config struct {
	AwsAcessKey         string  `key:"aws_access_key"`
	AwsSecretAccessKey  string  `key:"aws_secret_access_key"`
	AwsRegion           string  `key:"aws_region"`
	ServerCerticateName string  `key:"server_certificate_name"`
	Critical            float64 `key:"critical"`
	Warning             float64 `key:"warning"`
	Tag                 string  `key:"tag"`
}

func Flags(conf *Config) []cli.Flag {
	return []cli.Flag{
		cli.StringFlag{
			Name:        "aws_access_key",
			Value:       "AKIAIOSFODNN7EXAMPLE",
			Usage:       "AWS Access Key. Either set ENV['AWS_ACCESS_KEY'] or provide it as an option. Uses Default Credential if none are passed",
			EnvVar:      "AWS_ACCESS_KEY",
			Destination: &conf.AwsAcessKey,
		},
		cli.StringFlag{
			Name:        "aws_secret_access_key",
			Value:       "wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY",
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
		cli.Float64Flag{
			Name:        "critical",
			Value:       20,
			Usage:       "Trigger a critical when value is below VALUE",
			EnvVar:      "CRITICAL",
			Destination: &conf.Critical,
		},
		cli.Float64Flag{
			Name:        "warning",
			Value:       20,
			Usage:       "Trigger a warning when value is below VALUE",
			EnvVar:      "WARNING",
			Destination: &conf.Warning,
		},
		cli.StringFlag{
			Name:        "tag",
			Value:       "NAME",
			Usage:       "Add instance TAG value to warn/critical message.",
			EnvVar:      "TAG",
			Destination: &conf.Tag,
		},
	}
}
