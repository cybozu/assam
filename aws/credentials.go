package aws

import (
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/service/sts"
	"gopkg.in/ini.v1"
)

// SaveCredentials saves credentials to AWS credentials file.
func SaveCredentials(profileName string, credentials sts.Credentials) error {
	file := defaults.SharedCredentialsFilename()
	c, err := ini.Load(file)
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
