module github.com/cybozu/assam

go 1.20

require (
	github.com/aws/aws-sdk-go v1.54.11
	github.com/chromedp/cdproto v0.0.0-20240626232640-f933b107c653
	github.com/chromedp/chromedp v0.9.5
	github.com/google/uuid v1.6.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.8.1
	github.com/stretchr/testify v1.9.0
	gopkg.in/ini.v1 v1.67.0
)

require (
	github.com/chromedp/sysutil v1.0.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.3.2 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	golang.org/x/sys v0.16.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

// Exclude x/text affected by CVE-2020-28852.
// https://github.com/golang/go/issues/42536
exclude golang.org/x/text v0.15.0
