package idp

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"net/url"
	"time"
)

const (
	loginURLTemplate = "https://login.microsoftonline.com/%s/saml2?SAMLRequest=%s"
)

// Azure provides functionality of AzureAD as IdP
type Azure struct {
}

// Authenticate sends SAML request to Azure and fetches SAML response
func (a *Azure) Authenticate(samlRequest string, tenantID string) (string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := chromedp.New(ctx)
	if err != nil {
		return "", err
	}

	err = c.Run(ctx, network.Enable())
	if err != nil {
		return "", err
	}

	loginURL := fmt.Sprintf(loginURLTemplate, tenantID, url.QueryEscape(string(samlRequest)))
	err = c.Run(ctx, chromedp.Navigate(loginURL))
	if err != nil {
		return "", err
	}

	err = c.Run(ctx, chromedp.ActionFunc(func(_ context.Context, h cdp.Executor) error {
		for {
			time.Sleep(time.Second * 60)
		}
	}))
	if err != nil {
		return "", err
	}

	return "", nil
}
