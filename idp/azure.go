package idp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/chromedp/cdproto"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/cybozu/arws/aws"
	"net/url"
)

const (
	loginURLTemplate = "https://login.microsoftonline.com/%s/saml2?SAMLRequest=%s"
)

// Azure provides functionality of AzureAD as IdP
type Azure struct {
	samlRequest string
	tenantID    string
	msgChan     chan cdproto.Message
}

// NewAzure returns Azure
func NewAzure(samlRequest string, tenantID string) Azure {
	return Azure{
		samlRequest: samlRequest,
		tenantID:    tenantID,
		msgChan:     make(chan cdproto.Message),
	}
}

// Authenticate sends SAML request to Azure and fetches SAML response
func (a *Azure) Authenticate(ctx context.Context) (string, error) {
	c, err := a.setupCDP(ctx)
	if err != nil {
		return "", err
	}

	err = a.navigateToLoginURL(ctx, c)
	if err != nil {
		return "", err
	}

	response, err := a.fetchSAMLResponse(ctx, c)
	if err != nil {
		return "", err
	}

	err = a.shutdown(ctx, c)
	if err != nil {
		return "", err
	}

	return response, nil
}

func (a *Azure) logHandler(_ string, is ...interface{}) {
	go func() {
		for _, elem := range is {
			var msg cdproto.Message
			err := json.Unmarshal([]byte(fmt.Sprintf("%s", elem)), &msg)
			if err == nil {
				a.msgChan <- msg
			}
		}
	}()
}

func (a *Azure) setupCDP(ctx context.Context) (*chromedp.CDP, error) {
	// Need log handler to handle network events.
	c, err := chromedp.New(ctx, chromedp.WithLog(a.logHandler))
	if err != nil {
		return nil, err
	}

	err = c.Run(ctx, network.Enable())
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (a *Azure) navigateToLoginURL(ctx context.Context, c *chromedp.CDP) error {
	loginURL := fmt.Sprintf(loginURLTemplate, a.tenantID, url.QueryEscape(string(a.samlRequest)))
	return c.Run(ctx, chromedp.Navigate(loginURL))
}

func (a *Azure) fetchSAMLResponse(ctx context.Context, c *chromedp.CDP) (string, error) {
	var resp string
	err := c.Run(ctx, chromedp.ActionFunc(func(ctx context.Context, h cdp.Executor) error {
		for {
			var msg cdproto.Message

			select {
			case <-ctx.Done():
				return ctx.Err()
			case msg = <-a.msgChan:
			}

			switch msg.Method.String() {
			case "Network.requestWillBeSent":
				var req network.EventRequestWillBeSent
				err := json.Unmarshal(msg.Params, &req)
				if err != nil {
					return err
				}

				if req.Request.URL != aws.EndpointURL {
					continue
				}

				form, err := url.ParseQuery(req.Request.PostData)
				if err != nil {
					return err
				}

				samlResponse, ok := form["SAMLResponse"]
				if !ok || len(a.samlRequest) == 0 {
					return errors.New("no such key: SAMLResponse")
				}

				resp = samlResponse[0]

				return nil
			}
		}
	}))
	if err != nil {
		return "", err
	}

	return resp, nil
}

func (a *Azure) shutdown(ctx context.Context, c *chromedp.CDP) error {
	err := c.Shutdown(ctx)
	if err != nil {
		return err
	}

	err = c.Wait()
	if err != nil {
		return err
	}

	return nil
}
