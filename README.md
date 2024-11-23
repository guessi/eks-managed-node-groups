# eks-managed-node-groups

[![GitHub Actions](https://github.com/guessi/eks-managed-node-groups/actions/workflows/go.yml/badge.svg?branch=master)](https://github.com/guessi/eks-managed-node-groups/actions/workflows/go.yml)
[![GoDoc](https://godoc.org/github.com/guessi/eks-managed-node-groups?status.svg)](https://godoc.org/github.com/guessi/eks-managed-node-groups)
[![Go Report Card](https://goreportcard.com/badge/github.com/guessi/eks-managed-node-groups)](https://goreportcard.com/report/github.com/guessi/eks-managed-node-groups)
[![GitHub release](https://img.shields.io/github/release/guessi/eks-managed-node-groups.svg)](https://github.com/guessi/eks-managed-node-groups/releases/latest)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/guessi/eks-managed-node-groups)](https://github.com/guessi/eks-managed-node-groups/blob/master/go.mod)

## Usage

```bash
$ eks-managed-node-groups --help

NAME:
   eks-managed-node-groups - managed Amazon EKS node groups made easy

USAGE:
   eks-managed-node-groups [global options] command [command options]

VERSION:
   v1.0.0

COMMANDS:
   version, v  Print version number
   help, h     Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

## Sample Output

[![asciicast](https://asciinema.org/a/YaLICvOSkcIkxEqYRoAXjwP6l.svg)](https://asciinema.org/a/YaLICvOSkcIkxEqYRoAXjwP6l)

## Install

### Homebrew

```bash
brew tap guessi/tap && brew update && brew install eks-managed-node-groups
```

### For non-Homebrew users

<details><!-- markdownlint-disable-line -->
<summary>Click to expand!</summary><!-- markdownlint-disable-line -->

### For Linux users

```bash
curl -fsSL https://github.com/guessi/eks-managed-node-groups/releases/latest/download/eks-managed-node-groups-Linux-$(uname -m).tar.gz -o - | tar zxvf -
mv -vf ./eks-managed-node-groups /usr/local/bin/eks-managed-node-groups
```

### For macOS users

```bash
curl -fsSL https://github.com/guessi/eks-managed-node-groups/releases/latest/download/eks-managed-node-groups-Darwin-$(uname -m).tar.gz -o - | tar zxvf -
mv -vf ./eks-managed-node-groups /usr/local/bin/eks-managed-node-groups
```

### For Windows users

```powershell
$SRC = 'https://github.com/guessi/eks-managed-node-groups/releases/latest/download/eks-managed-node-groups-Windows-x86_64.tar.gz'
$DST = 'C:\Temp\eks-managed-node-groups-Windows-x86_64.tar.gz'
Invoke-RestMethod -Uri $SRC -OutFile $DST
```

</details>

## License

[Apache-2.0](LICENSE)
