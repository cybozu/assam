package main

import (
	"github.com/aws/aws-sdk-go/aws/defaults"
	"gopkg.in/ini.v1"
)

// Config is this tool's configuration
type Config struct {
	AppIDURI      string
	AzureTenantID string
}

const (
	appIDURIKeyName      = "app_id_uri"
	azureTenantIDKeyName = "azure_tenant_id"
)

// NewConfig returns Config from default AWS config file
func NewConfig() (Config, error) {
	cfg := Config{}

	file := defaults.SharedConfigFilename()
	f, err := ini.Load(file)
	if err != nil {
		return cfg, err
	}

	section, err := f.GetSection("default")
	if err != nil {
		return cfg, err
	}

	appIDURIKey, err := section.GetKey(appIDURIKeyName)
	if err != nil {
		return cfg, err
	}

	azureTenantIDKey, err := section.GetKey(azureTenantIDKeyName)
	if err != nil {
		return cfg, err
	}

	cfg.AppIDURI = appIDURIKey.Value()
	cfg.AzureTenantID = azureTenantIDKey.Value()

	return cfg, nil
}
