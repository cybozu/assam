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
	"path/filepath"
	"strconv"
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

	cmd := &cobra.Command{
		Use:   "assam",
		Short: "assam simplifies AssumeRoleWithSAML with CLI",
		Long:  `It is difficult to get a credential of AWS when using AssumeRoleWithSAML. This tool simplifies it.`,
		RunE: func(_ *cobra.Command, args []string) error {
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

	return cmd
}

func configureSettings(profile string) error {
	p := prompt.NewPrompt()
	cfg := config.Config{}

	var err error
	cfg.AzureTenantID, err = p.AskString("Azure Tenant ID", nil)
	if err != nil {
		return err
	}

	cfg.AppIDURI, err = p.AskString("App ID URI", nil)
	if err != nil {
		return err
	}

	cfg.DefaultSessionDurationHours, err = p.AskInt("Default Session Duration Hours (1-12)", &prompt.Options{
		ValidateFunc: func(val string) error {
			duration, err := strconv.Atoi(val)
			if err != nil || duration < 1 || 12 < duration {
				return fmt.Errorf("default session duration hours must be between 1 and 12: %s", val)
			}
			return nil
		},
	})
	if err != nil {
		return err
	}

	cfg.ChromeUserDataDir, err = p.AskString("Chrome User Data Directory", &prompt.Options{
		Default: filepath.Join(defaults.UserHomeDir(), ".config", "assam", "chrome-user-data"),
	})
	if err != nil {
		return err
	}

	return config.Save(cfg, profile)
}
