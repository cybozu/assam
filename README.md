arws
====

It is difficult to get a credential of AWS when using AssumeRoleWithSAML. This tool simplifies it.

## Requirement

The following operating systems are supported:

- Windows
- macOS
- Linux

And Google Chrome is required.

## Usage

```
Usage: arws [options]

options:
  -c, -configure
    Configuration Mode
  -p, -profile string
    AWS profile name (default: "default")
```

## Install

```bash
$ go get -u github.com/cybozu/arws
```

## Contribution

1. Fork ([https://github.com/cybozu/arws](https://github.com/cybozu/arws))
2. Create a feature branch
3. Commit your changes
4. Rebase your local changes against the master branch
5. Create new Pull Request
6. Green CI Tests

## Licence

[MIT](https://github.com/cybozu/arws/blob/master/LICENSE)
