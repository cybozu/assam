package main

import (
	"github.com/cybozu/arws/aws"
	"github.com/cybozu/arws/idp"
	"log"
)

func main() {
	cfg, err := NewConfig()
	if err != nil {
		log.Panic(err)
	}

	request, err := aws.CreateSAMLRequest(cfg.AppIDURI)
	if err != nil {
		log.Panic(err)
	}
	azure := idp.Azure{}
	response, err := azure.Authenticate(request, cfg.AzureTenantID)
	if err != nil {
		log.Panic(err)
	}
	log.Println(response)
}
