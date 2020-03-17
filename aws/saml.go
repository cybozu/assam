// Package aws provides the functionality about AWS.
package aws

import (
	"bytes"
	"compress/flate"
	"context"
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/google/uuid"
	"strings"
	"time"
)

const (
	// EndpointURL receives SAML response.
	EndpointURL = "https://signin.aws.amazon.com/saml"

	roleAttributeName = "https://aws.amazon.com/SAML/Attributes/Role"
)

// SAMLResponse is SAML response
type SAMLResponse struct {
	Assertion Assertion
}

// Assertion is an Assertion element of SAML response
type Assertion struct {
	AttributeStatement AttributeStatement
}

// AttributeStatement is an AttributeStatement element of SAML response
type AttributeStatement struct {
	Attributes []Attribute `xml:"Attribute"`
}

// Attribute is an Attribute element of SAML response
type Attribute struct {
	Name            string           `xml:",attr"`
	AttributeValues []AttributeValue `xml:"AttributeValue"`
}

// AttributeValue is an AttributeValue element of SAML response
type AttributeValue struct {
	Value string `xml:",innerxml"`
}

// CreateSAMLRequest creates the Base64 encoded SAML authentication request XML compressed by Deflate.
func CreateSAMLRequest(appIDURI string) (string, error) {
	// https://docs.microsoft.com/en-us/azure/active-directory/develop/single-sign-on-saml-protocol
	// ID must not begin with a number, so a common strategy is to prepend a string like "id" to the string
	// representation of a GUID.
	// See https://www.w3.org/TR/xmlschema-2/#ID
	xml := `
<samlp:AuthnRequest
  AssertionConsumerServiceURL="%s"
  ID="id_%s"
  IssueInstant="%s"
  ProtocolBinding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST"
  Version="2.0"
  xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol">
  <saml:Issuer xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion">%s</saml:Issuer>
  <samlp:NameIDPolicy Format="urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress" />
</samlp:AuthnRequest>
`

	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}

	instant := time.Now().Format(time.RFC3339)
	request := fmt.Sprintf(xml, EndpointURL, id, instant, appIDURI)

	deflated, err := deflate(request)
	if err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(deflated.Bytes())

	return encoded, nil
}

// ParseSAMLResponse parses base64 encoded response to SAMLResponse structure
func ParseSAMLResponse(base64Response string) (*SAMLResponse, error) {
	responseData, err := base64.StdEncoding.DecodeString(base64Response)
	if err != nil {
		return nil, err
	}

	response := SAMLResponse{}
	err = xml.Unmarshal(responseData, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// ExtractRoleArnAndPrincipalArn extracts role ARN and principal ARN from SAML response
func ExtractRoleArnAndPrincipalArn(samlResponse SAMLResponse, roleName string) (string, string, error) {
	for _, attr := range samlResponse.Assertion.AttributeStatement.Attributes {
		if attr.Name != roleAttributeName {
			continue
		}

		for _, v := range attr.AttributeValues {
			s := strings.Split(v.Value, ",")
			roleArn := s[0]
			principalArn := s[1]
			if roleName != "" && strings.Split(roleArn, "/")[1] != roleName {
				continue
			}
			return roleArn, principalArn, nil
		}
	}

	return "", "", fmt.Errorf("no such attribute: %s", roleAttributeName)
}

// AssumeRoleWithSAML sends a AssumeRoleWithSAML request to AWS and returns credentials
func AssumeRoleWithSAML(ctx context.Context, durationHours int, roleArn string, principalArn string, base64Response string) (*sts.Credentials, error) {
	sess := session.Must(session.NewSession())
	svc := sts.New(sess)

	input := sts.AssumeRoleWithSAMLInput{
		DurationSeconds: aws.Int64(int64(durationHours) * 60 * 60),
		RoleArn:         aws.String(roleArn),
		PrincipalArn:    aws.String(principalArn),
		SAMLAssertion:   aws.String(base64Response),
	}
	res, err := svc.AssumeRoleWithSAMLWithContext(ctx, &input)
	if err != nil {
		return nil, err
	}

	return res.Credentials, nil
}

func deflate(src string) (*bytes.Buffer, error) {
	b := new(bytes.Buffer)

	w, err := flate.NewWriter(b, 9)
	if err != nil {
		return nil, err
	}

	if _, err := w.Write([]byte(src)); err != nil {
		return nil, err
	}

	err = w.Close()
	if err != nil {
		return nil, err
	}

	return b, nil
}
