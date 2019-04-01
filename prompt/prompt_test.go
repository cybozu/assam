package prompt

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"testing"
)

func TestPrompt_AskString(t *testing.T) {
	type fields struct {
		writer  io.Writer
		scanner *bufio.Scanner
	}
	type args struct {
		query   string
		options *Options
	}
	type want struct {
		ret    string
		prompt string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "returns input value without options",
			fields: fields{
				writer:  new(bytes.Buffer),
				scanner: bufio.NewScanner(bytes.NewBufferString("hoge")),
			},
			args: args{
				query: "Please input something",
			},
			want: want{
				ret:    "hoge",
				prompt: "Please input something: ",
			},
		},
		{
			name: "returns trimmed input value without options",
			fields: fields{
				writer:  new(bytes.Buffer),
				scanner: bufio.NewScanner(bytes.NewBufferString(" hoge ")),
			},
			args: args{
				query: "Please input something",
			},
			want: want{
				ret:    "hoge",
				prompt: "Please input something: ",
			},
		},
		{
			name: "returns default value when input is empty",
			fields: fields{
				writer:  new(bytes.Buffer),
				scanner: bufio.NewScanner(bytes.NewBufferString("")),
			},
			args: args{
				query: "Please input something",
				options: &Options{
					Default: "default value",
				},
			},
			want: want{
				ret:    "default value",
				prompt: "Please input something (Default: default value): ",
			},
		},
		{
			name: "returns valid value when ValidateFunc is specified",
			fields: fields{
				writer:  new(bytes.Buffer),
				scanner: bufio.NewScanner(bytes.NewBufferString("invalid\nvalid")),
			},
			args: args{
				query: "Please input something",
				options: &Options{
					ValidateFunc: func(s string) error {
						if s != "valid" {
							return fmt.Errorf("invalid value: %s", s)
						}
						return nil
					},
				},
			},
			want: want{
				ret:    "valid",
				prompt: "Please input something: invalid value: invalid\nPlease input something: ",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Prompt{
				writer:  tt.fields.writer,
				scanner: tt.fields.scanner,
			}
			got, err := p.AskString(tt.args.query, tt.args.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("Prompt.AskString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want.ret {
				t.Errorf("Prompt.AskString() = '%v', want '%v'", got, tt.want.ret)
			}
			actualPrompt := tt.fields.writer.(*bytes.Buffer).Bytes()
			expectedPrompt := []byte(tt.want.prompt)
			if bytes.Compare(actualPrompt, expectedPrompt) != 0 {
				t.Errorf("Prompt of Prompt.AskString() = '%s', want '%s'", actualPrompt, expectedPrompt)
			}
		})
	}
}

func TestPrompt_AskInt(t *testing.T) {
	type fields struct {
		writer  io.Writer
		scanner *bufio.Scanner
	}
	type args struct {
		query   string
		options *Options
	}
	type want struct {
		ret    int
		prompt string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    want
		wantErr bool
	}{
		{
			name: "returns input value without options",
			fields: fields{
				writer:  new(bytes.Buffer),
				scanner: bufio.NewScanner(bytes.NewBufferString("123")),
			},
			args: args{
				query: "Please input something",
			},
			want: want{
				ret:    123,
				prompt: "Please input something: ",
			},
		},
		{
			name: "returns trimmed input value without options",
			fields: fields{
				writer:  new(bytes.Buffer),
				scanner: bufio.NewScanner(bytes.NewBufferString(" 123 ")),
			},
			args: args{
				query: "Please input something",
			},
			want: want{
				ret:    123,
				prompt: "Please input something: ",
			},
		},
		{
			name: "returns default value when input is empty",
			fields: fields{
				writer:  new(bytes.Buffer),
				scanner: bufio.NewScanner(bytes.NewBufferString("")),
			},
			args: args{
				query: "Please input something",
				options: &Options{
					Default: "999",
				},
			},
			want: want{
				ret:    999,
				prompt: "Please input something (Default: 999): ",
			},
		},
		{
			name: "returns valid value when ValidateFunc is specified",
			fields: fields{
				writer:  new(bytes.Buffer),
				scanner: bufio.NewScanner(bytes.NewBufferString("99\n100")),
			},
			args: args{
				query: "Please input something",
				options: &Options{
					ValidateFunc: func(s string) error {
						if s != "100" {
							return fmt.Errorf("invalid value: %s", s)
						}
						return nil
					},
				},
			},
			want: want{
				ret:    100,
				prompt: "Please input something: invalid value: 99\nPlease input something: ",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Prompt{
				writer:  tt.fields.writer,
				scanner: tt.fields.scanner,
			}
			got, err := p.AskInt(tt.args.query, tt.args.options)
			if (err != nil) != tt.wantErr {
				t.Errorf("Prompt.AskString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want.ret {
				t.Errorf("Prompt.AskString() = '%v', want '%v'", got, tt.want.ret)
			}
			actualPrompt := tt.fields.writer.(*bytes.Buffer).Bytes()
			expectedPrompt := []byte(tt.want.prompt)
			if bytes.Compare(actualPrompt, expectedPrompt) != 0 {
				t.Errorf("Prompt of Prompt.AskString() = '%s', want '%s'", actualPrompt, expectedPrompt)
			}
		})
	}
}
