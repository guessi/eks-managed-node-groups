# eks-managed-node-groups

[![GitHub Actions](https://github.com/guessi/eks-managed-node-groups/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/guessi/eks-managed-node-groups/actions/workflows/go.yml)
[![GoDoc](https://godoc.org/github.com/guessi/eks-managed-node-groups?status.svg)](https://godoc.org/github.com/guessi/eks-managed-node-groups)
[![Go Report Card](https://goreportcard.com/badge/github.com/guessi/eks-managed-node-groups)](https://goreportcard.com/report/github.com/guessi/eks-managed-node-groups)
[![GitHub release](https://img.shields.io/github/release/guessi/eks-managed-node-groups.svg)](https://github.com/guessi/eks-managed-node-groups/releases/latest)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/guessi/eks-managed-node-groups)](https://github.com/guessi/eks-managed-node-groups/blob/main/go.mod)

## Usage

![image](https://github.com/user-attachments/assets/f2ea8a99-44d6-4641-a937-a8ec7eb8ca4c)
![image](https://github.com/user-attachments/assets/8e324eac-2f0a-4a42-a7a9-2e9140ac7ff6)
![image](https://github.com/user-attachments/assets/2839e7f4-bba3-4273-99bd-54041f4c7451)
![image](https://github.com/user-attachments/assets/9e3f2d12-c697-4f61-80f2-5b9468ac25a0)

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
