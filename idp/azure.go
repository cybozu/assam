package idp

import (
	"context"
	"fmt"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/cybozu/assam/aws"
	"github.com/pkg/errors"
	"net/url"
	"os"
)

const (
	loginURLTemplate = "https://login.microsoftonline.com/%s/saml2?SAMLRequest=%s"
)

// Azure provides functionality of AzureAD as IdP
type Azure struct {
	samlRequest string
	tenantID    string
	msgChan     chan *network.EventRequestWillBeSent
}

// NewAzure returns Azure
func NewAzure(samlRequest string, tenantID string) Azure {
	return Azure{
		samlRequest: samlRequest,
		tenantID:    tenantID,
		msgChan:     make(chan *network.EventRequestWillBeSent),
	}
}

// Authenticate sends SAML request to Azure and fetches SAML response
func (a *Azure) Authenticate(ctx context.Context, userDataDir string) (string, error) {
	ctx, cancel := a.setupContext(ctx, userDataDir)
	defer cancel()

	// Need network.Enable() to handle network events.
	err := chromedp.Run(ctx, network.Enable())
	if err != nil {
		return "", err
	}

	a.listenNetworkRequest(ctx)

	err = a.navigateToLoginURL(ctx)
	if err != nil {
		return "", err
	}

	response, err := a.fetchSAMLResponse(ctx)
	if err != nil {
		return "", err
	}

	// Shut down gracefully to ensure that user data is stored.
	err = chromedp.Cancel(ctx)
	if err != nil {
		return "", err
	}

	return response, nil
}

func (a *Azure) setupContext(ctx context.Context, userDataDir string) (context.Context, context.CancelFunc) {
	// Need to expand environment variables because chromedp does not expand.
	expandedDir := os.ExpandEnv(userDataDir)

	opts := []chromedp.ExecAllocatorOption{
		chromedp.UserDataDir(expandedDir),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
	}

	allocContext, _ := chromedp.NewExecAllocator(context.Background(), opts...)

	return chromedp.NewContext(allocContext)
}

func (a *Azure) listenNetworkRequest(ctx context.Context) {
	chromedp.ListenTarget(ctx, func(v interface{}) {
		go func() {
			if req, ok := v.(*network.EventRequestWillBeSent); ok {
				a.msgChan <- req
			}
		}()
	})
}

func (a *Azure) navigateToLoginURL(ctx context.Context) error {
	loginURL := fmt.Sprintf(loginURLTemplate, a.tenantID, url.QueryEscape(string(a.samlRequest)))
	return chromedp.Run(ctx, chromedp.Navigate(loginURL))
}

func (a *Azure) fetchSAMLResponse(ctx context.Context) (string, error) {
	for {
		var req *network.EventRequestWillBeSent
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case req = <-a.msgChan:
		}

		if req.Request.URL != aws.EndpointURL {
			continue
		}

		form, err := url.ParseQuery(req.Request.PostData)
		if err != nil {
			return "", err
		}

		samlResponse, ok := form["SAMLResponse"]
		if !ok || len(a.samlRequest) == 0 {
			return "", errors.New("no such key: SAMLResponse")
		}

		return samlResponse[0], nil
	}
}
