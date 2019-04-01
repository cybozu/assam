package config

import (
	"github.com/aws/aws-sdk-go/aws/defaults"
	"gopkg.in/ini.v1"
	"strconv"
)

// Config is this tool's configuration
type Config struct {
	AppIDURI                    string
	AzureTenantID               string
	DefaultSessionDurationHours int
	ChromeUserDataDir           string
}

const (
	appIDURIKeyName                    = "app_id_uri"
	azureTenantIDKeyName               = "azure_tenant_id"
	defaultSessionDurationHoursKeyName = "default_session_duration_hours"
	chromeUserDataDirKeyName           = "chrome_user_data_dir"
)

// NewConfig returns Config from default AWS config file
func NewConfig() (Config, error) {
	cfg := Config{}

	f, err := loadConfigFile()
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

	defaultSessionDurationHoursKey, err := section.GetKey(defaultSessionDurationHoursKeyName)
	if err != nil {
		return cfg, err
	}

	userDataDirKey, err := section.GetKey(chromeUserDataDirKeyName)
	if err != nil {
		return cfg, err
	}

	cfg.AppIDURI = appIDURIKey.Value()
	cfg.AzureTenantID = azureTenantIDKey.Value()
	defaultSessionDurationHours, err := strconv.Atoi(defaultSessionDurationHoursKey.Value())
	if err != nil {
		return cfg, err
	}
	cfg.DefaultSessionDurationHours = defaultSessionDurationHours
	cfg.ChromeUserDataDir = userDataDirKey.Value()

	return cfg, nil
}

// Save saves config to file.
func Save(cfg Config) error {
	f, err := loadConfigFile()
	if err != nil {
		return err
	}

	section, err := f.GetSection("default")
	if err != nil {
		return err
	}

	section.Key(appIDURIKeyName).SetValue(cfg.AppIDURI)
	section.Key(azureTenantIDKeyName).SetValue(cfg.AzureTenantID)
	section.Key(defaultSessionDurationHoursKeyName).SetValue(strconv.Itoa(cfg.DefaultSessionDurationHours))
	section.Key(chromeUserDataDirKeyName).SetValue(cfg.ChromeUserDataDir)

	file := defaults.SharedConfigFilename()
	return f.SaveTo(file)
}

func loadConfigFile() (*ini.File, error) {
	file := defaults.SharedConfigFilename()
	return ini.Load(file)
}
