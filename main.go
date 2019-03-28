package main

import (
	"context"
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	azure := idp.NewAzure(request, cfg.AzureTenantID)
	base64Response, err := azure.Authenticate(ctx)
	if err != nil {
		log.Panic(err)
	}

	response, err := aws.ParseSAMLResponse(base64Response)
	if err != nil {
		log.Panic(err)
	}

	roleArn, principalArn, err := aws.ExtractRoleArnAndPrincipalArn(*response)
	if err != nil {
		log.Panic(err)
	}

	credentials, err := aws.AssumeRoleWithSAML(ctx, roleArn, principalArn, base64Response)
	if err != nil {
		log.Panic(err)
	}
	log.Println(credentials)
}
