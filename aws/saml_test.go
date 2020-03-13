package aws

import (
	"bytes"
	"compress/flate"
	"encoding/base64"
	"encoding/xml"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestParseSAMLResponse(t *testing.T) {
	t.Run("Parse SAML response", func(t *testing.T) {
		// setup
		response := `
<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" ID="_8e8dc5f69a98cc4c1ff3427e5ce34606fd672f91e6" Version="2.0" IssueInstant="2014-07-17T01:01:48Z" Destination="http://sp.example.com/demo1/index.php?acs" InResponseTo="ONELOGIN_4fee3b046395c4e751011e97f8900b5273d56685">
  <saml:Issuer>http://idp.example.com/metadata.php</saml:Issuer>
  <samlp:Status>
    <samlp:StatusCode Value="urn:oasis:names:tc:SAML:2.0:status:Success"/>
  </samlp:Status>
  <saml:Assertion xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xmlns:xs="http://www.w3.org/2001/XMLSchema" ID="_d71a3a8e9fcc45c9e9d248ef7049393fc8f04e5f75" Version="2.0" IssueInstant="2014-07-17T01:01:48Z">
    <saml:Issuer>http://idp.example.com/metadata.php</saml:Issuer>
    <saml:Subject>
      <saml:NameID SPNameQualifier="http://sp.example.com/demo/metadata" Format="urn:oasis:names:tc:SAML:2.0:nameid-format:transient">_ce3d2948b4cf20146dee0a0b3dd6f69b6cf86f62d7</saml:NameID>
      <saml:SubjectConfirmation Method="urn:oasis:names:tc:SAML:2.0:cm:bearer">
        <saml:SubjectConfirmationData NotOnOrAfter="2024-01-18T06:21:48Z" Recipient="http://sp.example.com/demo1/index.php?acs" InResponseTo="ONELOGIN_4fee3b046395c4e751011e97f8900b5273d56685"/>
      </saml:SubjectConfirmation>
    </saml:Subject>
    <saml:Conditions NotBefore="2014-07-17T01:01:18Z" NotOnOrAfter="2024-01-18T06:21:48Z">
      <saml:AudienceRestriction>
        <saml:Audience>http://sp.example.com/demo1/metadata.php</saml:Audience>
      </saml:AudienceRestriction>
    </saml:Conditions>
    <saml:AuthnStatement AuthnInstant="2014-07-17T01:01:48Z" SessionNotOnOrAfter="2024-07-17T09:01:48Z" SessionIndex="_be9967abd904ddcae3c0eb4189adbe3f71e327cf93">
      <saml:AuthnContext>
        <saml:AuthnContextClassRef>urn:oasis:names:tc:SAML:2.0:ac:classes:Password</saml:AuthnContextClassRef>
      </saml:AuthnContext>
    </saml:AuthnStatement>
    <saml:AttributeStatement>
      <saml:Attribute Name="uid" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic">
        <saml:AttributeValue xsi:type="xs:string">test</saml:AttributeValue>
      </saml:Attribute>
      <saml:Attribute Name="mail" NameFormat="urn:oasis:names:tc:SAML:2.0:attrname-format:basic">
        <saml:AttributeValue xsi:type="xs:string">test@example.com</saml:AttributeValue>
      </saml:Attribute>
    </saml:AttributeStatement>
  </saml:Assertion>
</samlp:Response>
`
		base64Response := base64.StdEncoding.EncodeToString([]byte(response))

		// exercise
		got, err := ParseSAMLResponse(base64Response)
		if err != nil {
			t.Errorf("ParseSAMLResponse() error = %v", err)
			return
		}

		// verify
		assert.Equal(t, "uid", got.Assertion.AttributeStatement.Attributes[0].Name)
		assert.Equal(t, "test", got.Assertion.AttributeStatement.Attributes[0].AttributeValues[0].Value)
		assert.Equal(t, "mail", got.Assertion.AttributeStatement.Attributes[1].Name)
		assert.Equal(t, "test@example.com", got.Assertion.AttributeStatement.Attributes[1].AttributeValues[0].Value)
	})
}

func TestExtractRoleArnAndPrincipalArn(t *testing.T) {
	type args struct {
		samlResponse SAMLResponse
		roleName     string
	}
	tests := []struct {
		name             string
		args             args
		wantRoleArn      string
		wantPrincipalArn string
		wantErr          bool
	}{
		{
			name: "extracts role ARN and principal ARN",
			args: args{
				samlResponse: SAMLResponse{
					Assertion: Assertion{
						AttributeStatement: AttributeStatement{
							Attributes: []Attribute{
								{
									Name: "dummy",
									AttributeValues: []AttributeValue{
										{
											Value: "dummy",
										},
									},
								},
								{
									Name: roleAttributeName,
									AttributeValues: []AttributeValue{
										{
											Value: "arn:aws:iam::012345678901:role/TestRole,arn:aws:iam::012345678901:saml-provider/TestProvider",
										},
									},
								},
							},
						},
					},
				},
				roleName: "",
			},
			wantRoleArn:      "arn:aws:iam::012345678901:role/TestRole",
			wantPrincipalArn: "arn:aws:iam::012345678901:saml-provider/TestProvider",
		},
		{
			name: "returns first role when role attribute are multi and no roleName argument",
			args: args{
				samlResponse: SAMLResponse{
					Assertion: Assertion{
						AttributeStatement: AttributeStatement{
							Attributes: []Attribute{
								{
									Name: "dummy",
									AttributeValues: []AttributeValue{
										{
											Value: "dummy",
										},
									},
								},
								{
									Name: roleAttributeName,
									AttributeValues: []AttributeValue{
										{
											Value: "arn:aws:iam::012345678901:role/TestRole1,arn:aws:iam::012345678901:saml-provider/TestProvider1",
										},
										{
											Value: "arn:aws:iam::012345678901:role/TestRole2,arn:aws:iam::012345678901:saml-provider/TestProvider2",
										},
									},
								},
							},
						},
					},
				},
				roleName: "",
			},
			wantRoleArn:      "arn:aws:iam::012345678901:role/TestRole1",
			wantPrincipalArn: "arn:aws:iam::012345678901:saml-provider/TestProvider1",
		},
		{
			name: "returns specify role when role attribute are multi and roleName argument",
			args: args{
				samlResponse: SAMLResponse{
					Assertion: Assertion{
						AttributeStatement: AttributeStatement{
							Attributes: []Attribute{
								{
									Name: "dummy",
									AttributeValues: []AttributeValue{
										{
											Value: "dummy",
										},
									},
								},
								{
									Name: roleAttributeName,
									AttributeValues: []AttributeValue{
										{
											Value: "arn:aws:iam::012345678901:role/TestRole1,arn:aws:iam::012345678901:saml-provider/TestProvider1",
										},
										{
											Value: "arn:aws:iam::012345678901:role/TestRole2,arn:aws:iam::012345678901:saml-provider/TestProvider2",
										},
									},
								},
							},
						},
					},
				},
				roleName: "TestRole2",
			},
			wantRoleArn:      "arn:aws:iam::012345678901:role/TestRole2",
			wantPrincipalArn: "arn:aws:iam::012345678901:saml-provider/TestProvider2",
		},
		{
			name: "returns an error when role attribute does not exist",
			args: args{
				samlResponse: SAMLResponse{
					Assertion: Assertion{
						AttributeStatement: AttributeStatement{
							Attributes: []Attribute{
								{
									Name: "dummy",
									AttributeValues: []AttributeValue{
										{
											Value: "dummy",
										},
									},
								},
							},
						},
					},
				},
				roleName: "",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := ExtractRoleArnAndPrincipalArn(tt.args.samlResponse, tt.args.roleName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractRoleArnAndPrincipalArn() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.wantRoleArn {
				t.Errorf("ExtractRoleArnAndPrincipalArn() got = %v, wantRoleArn %v", got, tt.wantRoleArn)
			}
			if got1 != tt.wantPrincipalArn {
				t.Errorf("ExtractRoleArnAndPrincipalArn() got1 = %v, wantRoleArn %v", got1, tt.wantPrincipalArn)
			}
		})
	}
}
