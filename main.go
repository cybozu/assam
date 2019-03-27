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
	azure := idp.NewAzure(request, cfg.AzureTenantID)
	response, err := azure.Authenticate()
	if err != nil {
		log.Panic(err)
	}
	log.Println(response)
}
