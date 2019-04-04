package cmd

import (
	"context"
	"fmt"
	"github.com/cybozu/assam/aws"
	"github.com/cybozu/assam/config"
	"github.com/cybozu/assam/defaults"
	"github.com/cybozu/assam/idp"
	"github.com/cybozu/assam/prompt"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
)

// goreleaser embed variables by ldflags
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

// Execute runs root command
func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	var configure bool
	var profile string
	var showVersion bool

	cmd := &cobra.Command{
		Use:   "assam",
		Short: "assam simplifies AssumeRoleWithSAML with CLI",
		Long:  `It is difficult to get a credential of AWS when using AssumeRoleWithSAML. This tool simplifies it.`,
		RunE: func(_ *cobra.Command, args []string) error {
			if showVersion {
				printVersion()
				return nil
			}

			if configure {
				err := configureSettings(profile)
				if err != nil {
					return err
				}
				return nil
			}

			cfg, err := config.NewConfig(profile)
			if err != nil {
				return err
			}

			request, err := aws.CreateSAMLRequest(cfg.AppIDURI)
			if err != nil {
				return err
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			handleSignal(cancel)

			azure := idp.NewAzure(request, cfg.AzureTenantID)
			base64Response, err := azure.Authenticate(ctx, cfg.ChromeUserDataDir)
			if err != nil {
				return err
			}

			response, err := aws.ParseSAMLResponse(base64Response)
			if err != nil {
				return err
			}

			roleArn, principalArn, err := aws.ExtractRoleArnAndPrincipalArn(*response)
			if err != nil {
				return err
			}

			credentials, err := aws.AssumeRoleWithSAML(ctx, roleArn, principalArn, base64Response)
			if err != nil {
				return err
			}

			err = aws.SaveCredentials(profile, *credentials)
			if err != nil {
				return err
			}

			return nil
		},
	}
	cmd.PersistentFlags().BoolVarP(&configure, "configure", "c", false, "configure initial settings")
	cmd.PersistentFlags().StringVarP(&profile, "profile", "p", "default", "AWS profile")
	cmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "Show version")

	return cmd
}

func printVersion() {
	fmt.Printf("version: %s, commit: %s, date: %s\n", version, commit, date)
}

func configureSettings(profile string) error {
	p := prompt.NewPrompt()

	// Load current config.
	cfg, err := config.NewConfig(profile)
	if err != nil {
		cfg = config.Config{}
	}

	// Azure Tenant ID
	var azureTenantIDOptions prompt.Options
	if cfg.AzureTenantID != "" {
		azureTenantIDOptions.Default = cfg.AzureTenantID
	}
	cfg.AzureTenantID, err = p.AskString("Azure Tenant ID", &azureTenantIDOptions)
	if err != nil {
		return err
	}

	// App ID URI
	var appIDURIOptions prompt.Options
	if cfg.AppIDURI != "" {
		appIDURIOptions.Default = cfg.AppIDURI
	}
	cfg.AppIDURI, err = p.AskString("App ID URI", &appIDURIOptions)
	if err != nil {
		return err
	}

	// Default session duration hours
	var defaultSessionDurationHoursOptions prompt.Options
	if cfg.DefaultSessionDurationHours != 0 {
		defaultSessionDurationHoursOptions.Default = strconv.Itoa(cfg.DefaultSessionDurationHours)
	}
	defaultSessionDurationHoursOptions.ValidateFunc = func(val string) error {
		duration, err := strconv.Atoi(val)
		if err != nil || duration < 1 || 12 < duration {
			return fmt.Errorf("default session duration hours must be between 1 and 12: %s", val)
		}
		return nil
	}
	cfg.DefaultSessionDurationHours, err = p.AskInt("Default Session Duration Hours (1-12)", &defaultSessionDurationHoursOptions)
	if err != nil {
		return err
	}

	// Chrome user data directory
	var chromeUserDataDirOptions prompt.Options
	if cfg.ChromeUserDataDir != "" {
		chromeUserDataDirOptions.Default = cfg.ChromeUserDataDir
	} else {
		chromeUserDataDirOptions.Default = filepath.Join(defaults.UserHomeDir(), ".config", "assam", "chrome-user-data")
	}
	cfg.ChromeUserDataDir, err = p.AskString("Chrome User Data Directory", &chromeUserDataDirOptions)
	if err != nil {
		return err
	}

	return config.Save(cfg, profile)
}

func handleSignal(cancel context.CancelFunc) {
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	go func() {
		for {
			<-signalChan
			cancel()
		}
	}()
}
