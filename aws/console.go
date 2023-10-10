package aws

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
)

// AWSClient is an interface for AWS operations
type awsClientInterface interface {
	GetConsoleURL() (string, error)
}

// awsClient is the implementation of AWSClient interface
type awsClient struct {
	session *session.Session
}

// NewAWSClient creates a new AWSClient instance
//
//	By default NewSession will only load credentials from the shared credentials file (~/.aws/credentials).
func NewAWSClient() awsClientInterface {
	// Create session
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))

	return &awsClient{
		session: sess,
	}
}

// GetConsoleURL returns the AWS Management Console URL
// ref: https://docs.aws.amazon.com/IAM/latest/UserGuide/id_roles_providers_enable-console-custom-url.html
func (c *awsClient) GetConsoleURL() (string, error) {
	amazonDomain := c.getConsoleDomain(*c.session.Config.Region)

	// Create get signin token URL
	creds, err := c.session.Config.Credentials.Get()
	if err != nil {
		return "", errors.New("failed to get aws credential: please authenticate with `assam`")
	}

	token, err := c.getSigninToken(creds, amazonDomain)
	if err != nil {
		return "", err
	}

	targetURL := fmt.Sprintf("https://console.%s/console/home", amazonDomain)
	params := url.Values{
		"Action":      []string{"login"},
		"Destination": []string{targetURL},
		"SigninToken": []string{token},
	}

	return fmt.Sprintf("https://signin.%s/federation?%s", amazonDomain, params.Encode()), nil
}

// getConsoleDomain returns the console domain based on the region
func (c *awsClient) getConsoleDomain(region string) string {
	var amazonDomain string

	if strings.HasPrefix(region, "us-gov-") {
		amazonDomain = "amazonaws-us-gov.com"
	} else if strings.HasPrefix(region, "cn-") {
		amazonDomain = "amazonaws.cn"
	} else {
		amazonDomain = "aws.amazon.com"
	}
	return amazonDomain
}

// getSinginToken retrieves the signin token
func (c *awsClient) getSigninToken(creds credentials.Value, amazonDomain string) (string, error) {
	urlCreds := map[string]string{
		"sessionId":    creds.AccessKeyID,
		"sessionKey":   creds.SecretAccessKey,
		"sessionToken": creds.SessionToken,
	}

	bytes, err := json.Marshal(urlCreds)
	if err != nil {
		return "", err
	}

	params := url.Values{
		"Action":          []string{"getSigninToken"},
		"DurationSeconds": []string{"900"}, // DurationSeconds minimum value
		"Session":         []string{string(bytes)},
	}
	tokenRequest := fmt.Sprintf("https://signin.%s/federation?%s", amazonDomain, params.Encode())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Construct a request to the federation URL.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, tokenRequest, nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("request failed: %s", resp.Status)
	}

	// Extract a signin token from the response body.
	token, err := c.getToken(resp.Body)
	if err != nil {
		return "", err
	}

	return token, nil
}

// getToken extracts the signin token from the response body
func (c *awsClient) getToken(reader io.Reader) (string, error) {
	type response struct {
		SigninToken string
	}

	var resp response
	if err := json.NewDecoder(reader).Decode(&resp); err != nil {
		return "", err
	}

	return resp.SigninToken, nil
}
