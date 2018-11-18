package main

import (
	"github.com/sreejita-biswas/aws-plugins/config"
	"github.com/sreejita-biswas/aws-plugins/aws_session"
	"github.com/sreejita-biswas/aws-plugins/aws_clients"
	"github.com/sreejita-biswas/aws-plugins/plugins"
	"fmt"
	"os"
)

var (
	conf = config.Config{}
)

func main(){
	certificates := []plugins.IAMServerCertificate{}
	var err error
	//Create AWS Session
	aws_session.CreateAwsSession(conf)

	//Create IAM AWS Client
	aws_clients.NewIAM()

	//list Server Certificate Expiration Dates
	if conf.ServerCerticateName == ""{
		certificates,err = plugins.ListServerCertificates(nil)
	}else{
		certificates,err = plugins.ListServerCertificates(&conf.ServerCerticateName)
	}

	if err != nil{
		fmt.Println("Error while getting server certificates : ", err)
		os.Exit(0)
	}

	fmt.Println("Server Certificates : ")
	for _,certificate := range certificates{
		fmt.Println("---------------------------")
		fmt.Println("Name : ", certificate.Name)
		fmt.Println("Expiration : ", certificate.Expiration)
	}
	fmt.Println("---------------------------")
	os.Exit(0)
}
