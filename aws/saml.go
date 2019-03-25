// Package aws provides the functionality to send requests to AWS.
package aws

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"fmt"
	"github.com/google/uuid"
	"time"
)

// CreateSAMLRequest creates the Base64 encoded SAML authentication request XML that compressed by Deflate.
func CreateSAMLRequest(appIDURI string) (string, error) {
	xml := `
<samlp:AuthnRequest
  AssertionConsumerServiceURL="https://signin.aws.amazon.com/saml"
  ID="%s"
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
