package aws

import (
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/service/sts"
	"gopkg.in/ini.v1"
	"os"
)

// SaveCredentials saves credentials to AWS credentials file.
func SaveCredentials(profileName string, credentials sts.Credentials) error {
	file := getCredentialsFilename()
	c, err := ini.LooseLoad(file)
	if err != nil {
		return err
	}

	s := c.Section(profileName)
	s.Key("aws_access_key_id").SetValue(*credentials.AccessKeyId)
	s.Key("aws_secret_access_key").SetValue(*credentials.SecretAccessKey)
	s.Key("aws_session_token").SetValue(*credentials.SessionToken)
	s.Key("aws_session_expiration").SetValue(credentials.Expiration.String())

	return c.SaveTo(file)
}

func getCredentialsFilename() string {
	// https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-envvars.html
	file := os.Getenv("AWS_SHARED_CREDENTIALS_FILE")
	if len(file) == 0 {
		file = defaults.SharedCredentialsFilename()
	}
	return file
}
