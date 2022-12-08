// Package prompt provides CLI prompt.
package prompt

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

// Prompt provides CLI prompt functionality
type Prompt struct {
	writer  io.Writer
	scanner *bufio.Scanner
}

// Options is option for Prompt
type Options struct {
	Default      string
	ValidateFunc ValidateFunc
}

// ValidateFunc is function to validate input value
type ValidateFunc func(string) error

// NewPrompt returns Prompt struct
func NewPrompt() Prompt {
	return Prompt{
		writer:  os.Stdout,
		scanner: bufio.NewScanner(os.Stdin),
	}
}

// AskString asks query and returns input string
func (p *Prompt) AskString(query string, options *Options) (string, error) {
	if options == nil {
		options = &Options{}
	}

	for {
		// prompt
		err := p.printPrompt(query, options)
		if err != nil {
			return "", err
		}

		// scan
		val, err := p.scanString()
		if err != nil {
			return "", err
		}
		if val == "" {
			return options.Default, nil
		}

		// validate
		if options.ValidateFunc != nil {
			err := options.ValidateFunc(val)
			if err != nil {
				_, err := fmt.Fprintln(p.writer, err.Error())
				if err != nil {
					return "", err
				}
				continue
			}
		}

		return val, nil
	}
}

// AskInt asks query and returns input int
func (p *Prompt) AskInt(query string, options *Options) (int, error) {
	if options == nil {
		options = &Options{}
	}

	for {
		// prompt
		err := p.printPrompt(query, options)
		if err != nil {
			return 0, err
		}

		// scan
		val, err := p.scanString()
		if err != nil {
			return 0, err
		}
		if val == "" {
			return strconv.Atoi(options.Default)
		}

		// validate
		if options.ValidateFunc != nil {
			err := options.ValidateFunc(val)
			if err != nil {
				_, err := fmt.Fprintln(p.writer, err.Error())
				if err != nil {
					return 0, err
				}
				continue
			}
		}

		return strconv.Atoi(val)
	}
}

func (p *Prompt) printPrompt(query string, options *Options) error {
	if options.Default != "" {
		_, err := fmt.Fprintf(p.writer, "%s (Default: %s): ", query, options.Default)
		if err != nil {
			return err
		}
	} else {
		_, err := fmt.Fprintf(p.writer, "%s: ", query)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Prompt) scanString() (string, error) {
	if !p.scanner.Scan() {
		return "", p.scanner.Err()
	}

	return strings.TrimSpace(p.scanner.Text()), nil
}
