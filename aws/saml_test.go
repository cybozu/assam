package aws

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"encoding/xml"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"testing"
)

type SAMLRequest struct {
	XMLName                     xml.Name     `xml:"AuthnRequest"`
	XMLNamespace                string       `xml:"xmlns samlp,attr"`
	AssertionConsumerServiceURL string       `xml:"AssertionConsumerServiceURL,attr"`
	ID                          string       `xml:"ID,attr"`
	IssueInstant                string       `xml:"IssueInstant,attr"`
	ProtocolBinding             string       `xml:"ProtocolBinding,attr"`
	Version                     string       `xml:"Version,attr"`
	Issuer                      Issuer       `xml:"Issuer"`
	NameIDPolicy                NameIDPolicy `xml:"NameIDPolicy"`
}

type Issuer struct {
	XMLName      xml.Name `xml:"Issuer"`
	XMLNamespace string   `xml:"xmlns saml,attr"`
	AppIDURI     string   `xml:",chardata"`
}

type NameIDPolicy struct {
	XMLName xml.Name `xml:"NameIDPolicy"`
	Format  string   `xml:"Format,attr"`
}

func TestCreateSAMLRequest(t *testing.T) {
	t.Run("Should have App ID URI at Issuer element", func(t *testing.T) {
		// setup
		appIDURI := "https://signin.aws.amazon.com/saml#sample"

		// exercise
		got, err := CreateSAMLRequest(appIDURI)
		if err != nil {
			t.Errorf("CreateSAMLRequest() error = %v", err)
			return
		}

		// verify
		b, err := base64.StdEncoding.DecodeString(got)
		if err != nil {
			t.Error(err)
			return
		}

		r := flate.NewReader(bytes.NewReader(b))
		xmlData, err := ioutil.ReadAll(r)
		if err != nil {
			t.Error(err)
			return
		}

		request := SAMLRequest{}
		err = xml.Unmarshal(xmlData, &request)
		if err != nil {
			t.Error(err)
			return
		}

		assert.Equal(t, "https://signin.aws.amazon.com/saml", request.AssertionConsumerServiceURL)
		assert.Equal(t, "urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST", request.ProtocolBinding)
		assert.Equal(t, "2.0", request.Version)
		assert.Equal(t, "urn:oasis:names:tc:SAML:2.0:protocol", request.XMLNamespace)
		assert.Equal(t, "urn:oasis:names:tc:SAML:2.0:assertion", request.Issuer.XMLNamespace)
		assert.Equal(t, "urn:oasis:names:tc:SAML:1.1:nameid-format:emailAddress", request.NameIDPolicy.Format)

		assert.NotEmpty(t, request.ID)
		assert.NotEmpty(t, request.IssueInstant)
		assert.Equal(t, appIDURI, request.Issuer.AppIDURI)
	})
}
