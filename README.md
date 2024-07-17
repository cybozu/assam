assam
====

> [!WARNING]
> assam was deprecated on 31 July 2024. Installation of assam by Homebrew will be disabled on 31 July 2025.
> As an alternative, you can use the official [AWS IAM Identity Center](https://docs.aws.amazon.com/singlesignon/latest/userguide/what-is.html) or fork assam and maintain it yourself.

It is difficult to get a credential of AWS when using AssumeRoleWithSAML. This tool simplifies it.

## Requirement

The following operating systems are supported:

- Windows
- macOS
- Linux

And Google Chrome is required.

## Usage

```
Usage: assam [options]

options:
  -c, --configure
    Configuration Mode
  -p, --profile string
    AWS profile name (default: "default")
  -w, --web
    Open the AWS Console URL in your default browser (*1)
```

Please be careful that assam overrides default profile in `.aws/credentials` by default.
If you don't want that, please specify `-p|--profile` option.

## Install

### Homebrew

```bash
$ brew install cybozu/assam/assam
```

### Manual

Download a binary file from [Release](https://github.com/cybozu/assam/releases) and save it to the desired location.

## Notes

### (*1) Command to open the default browser

- Windows: `start`
- macOS : `open`
- Linux: `xdg-open`

## Contribution

1. Fork ([https://github.com/cybozu/assam](https://github.com/cybozu/assam))
2. Create a feature branch
3. Commit your changes
4. Rebase your local changes against the master branch
5. Create new Pull Request
6. Green CI Tests

## Licence

[MIT](https://github.com/cybozu/assam/blob/master/LICENSE)
