// Package aws provides the functionality about AWS.
package aws

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"time"
)

// CreateSAMLRequest creates the Base64 encoded SAML authentication request XML compressed by Deflate.
func CreateSAMLRequest(appIDURI string) (string, error) {
	// https://docs.microsoft.com/en-us/azure/active-directory/develop/single-sign-on-saml-protocol
	// ID must not begin with a number, so a common strategy is to prepend a string like "id" to the string
	// representation of a GUID.
	// See https://www.w3.org/TR/xmlschema-2/#ID
	xml := `
<samlp:AuthnRequest
  AssertionConsumerServiceURL="https://signin.aws.amazon.com/saml"
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
	request := fmt.Sprintf(xml, id, instant, appIDURI)

	deflated, err := deflate(request)
	if err != nil {
		return "", err
	}

	encoded := base64.StdEncoding.EncodeToString(deflated.Bytes())

	return encoded, nil
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
