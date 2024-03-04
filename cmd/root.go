// Package cmd provides assam CLI.
package cmd

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/cybozu/assam/aws"
	"github.com/cybozu/assam/config"
	"github.com/cybozu/assam/defaults"
	"github.com/cybozu/assam/idp"
	"github.com/cybozu/assam/prompt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"os"
	"os/exec"
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
		// Not print an error because cobra.Command prints it.
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	var configure bool
	var roleName string
	var profile string
	var web bool
	var showVersion bool

	cmd := &cobra.Command{
		Use:          "assam",
		Short:        "assam simplifies AssumeRoleWithSAML with CLI",
		Long:         `It is difficult to get a credential of AWS when using AssumeRoleWithSAML. This tool simplifies it.`,
		SilenceUsage: true,
		RunE: func(_ *cobra.Command, _ []string) error {
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

			if web {
				return openBrowser()
			}

			cfg, err := config.NewConfig(profile)
			if err != nil {
				return errors.Wrap(err, "please run `assam --configure` at the first time")
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

			roleArn, principalArn, err := aws.ExtractRoleArnAndPrincipalArn(*response, roleName)
			if err != nil {
				return err
			}

			credentials, err := aws.AssumeRoleWithSAML(ctx, cfg.DefaultSessionDurationHours, roleArn, principalArn, base64Response)
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
	cmd.PersistentFlags().StringVarP(&roleName, "role", "r", "", "AWS IAM role name")
	cmd.PersistentFlags().BoolVarP(&web, "web", "w", false, "open AWS management console in a browser")
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

func openBrowser() error {
	url, err := aws.NewAWSClient().GetConsoleURL()
	if err != nil {
		return err
	}

	var cmd string
	var args []string
	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", strings.ReplaceAll(url, "&", "^&")} // for Windows: "&! <>^|" etc. must be escaped, but since only "&" is used, the corresponding
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	}

	if len(cmd) != 0 {
		err = exec.Command(cmd, args...).Run()
		if err != nil {
			return err
		}
	} else {
		return errors.New("OS does not support -web command")
	}
	return nil
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
