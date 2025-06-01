# eks-managed-node-groups

[![GitHub Actions](https://github.com/guessi/eks-managed-node-groups/actions/workflows/go.yml/badge.svg?branch=main)](https://github.com/guessi/eks-managed-node-groups/actions/workflows/go.yml)
[![GoDoc](https://godoc.org/github.com/guessi/eks-managed-node-groups?status.svg)](https://godoc.org/github.com/guessi/eks-managed-node-groups)
[![Go Report Card](https://goreportcard.com/badge/github.com/guessi/eks-managed-node-groups)](https://goreportcard.com/report/github.com/guessi/eks-managed-node-groups)
[![GitHub release](https://img.shields.io/github/release/guessi/eks-managed-node-groups.svg)](https://github.com/guessi/eks-managed-node-groups/releases/latest)
[![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/guessi/eks-managed-node-groups)](https://github.com/guessi/eks-managed-node-groups/blob/main/go.mod)

## 🤔 Why we need this? what it is trying to resolve?

In the real world, not all [Amazon EKS](https://docs.aws.amazon.com/eks/latest/userguide/what-is-eks.html) clusters require [autosclaing](https://docs.aws.amazon.com/eks/latest/userguide/autoscaling.html) mechanism, some of them needs to be manually scaled. To reduce human errors, `eks-managed-node-groups` is here to solve this problem, `eks-managed-node-groups` provides an user friendly [TUI (Text-based User Interface)](https://en.wikipedia.org/wiki/Text-based_user_interface) allows users to easily jump between clusters with only a few options, without having to log into the Amazon EKS console and jump between configuration pages.

## 🔢 Prerequisites

* An existing [Amazon EKS](https://docs.aws.amazon.com/eks/latest/userguide/what-is-eks.html) cluster.
* An existing [kubeconfig](https://docs.aws.amazon.com/eks/latest/userguide/create-kubeconfig.html).
* [Grant IAM users and roles access to Kubernetes APIs](https://docs.aws.amazon.com/eks/latest/userguide/grant-k8s-access.html).
* An IAM Role/User with the following permissions:
    * [autoscaling:DescribeAutoScalingGroups](https://docs.aws.amazon.com/autoscaling/ec2/APIReference/API_DescribeAutoScalingGroups.html)
    * [autoscaling:SetDesiredCapacity](https://docs.aws.amazon.com/autoscaling/ec2/APIReference/API_SetDesiredCapacity.html)
    * [eks:DescribeNodegroup](https://docs.aws.amazon.com/eks/latest/APIReference/API_DescribeNodegroup.html)
    * [eks:ListNodegroups](https://docs.aws.amazon.com/eks/latest/APIReference/API_ListNodegroups.html)
    * [eks:UpdateNodegroupConfig](https://docs.aws.amazon.com/eks/latest/APIReference/API_UpdateNodegroupConfig.html)

## 🚀 Quick start

![image](https://github.com/user-attachments/assets/f2ea8a99-44d6-4641-a937-a8ec7eb8ca4c)
![image](https://github.com/user-attachments/assets/8e324eac-2f0a-4a42-a7a9-2e9140ac7ff6)
![image](https://github.com/user-attachments/assets/2839e7f4-bba3-4273-99bd-54041f4c7451)
![image](https://github.com/user-attachments/assets/9e3f2d12-c697-4f61-80f2-5b9468ac25a0)

## 👷 Install

### For macOS/Linux users (Recommended)

Brand new install

```bash
brew tap guessi/tap && brew update && brew install eks-managed-node-groups
```

To upgrade version

```bash
brew update && brew upgrade eks-managed-node-groups
```

### Manually setup (Linux, Windows, macOS)

<details><!-- markdownlint-disable-line -->
<summary>Click to expand!</summary><!-- markdownlint-disable-line -->

#### For Linux users

```bash
curl -fsSL https://github.com/guessi/eks-managed-node-groups/releases/latest/download/eks-managed-node-groups-Linux-$(uname -m).tar.gz -o - | tar zxvf -
mv -vf ./eks-managed-node-groups /usr/local/bin/eks-managed-node-groups
```

#### For macOS users

```bash
curl -fsSL https://github.com/guessi/eks-managed-node-groups/releases/latest/download/eks-managed-node-groups-Darwin-$(uname -m).tar.gz -o - | tar zxvf -
mv -vf ./eks-managed-node-groups /usr/local/bin/eks-managed-node-groups
```

#### For Windows users

```powershell
$SRC = 'https://github.com/guessi/eks-managed-node-groups/releases/latest/download/eks-managed-node-groups-Windows-x86_64.tar.gz'
$DST = 'C:\Temp\eks-managed-node-groups-Windows-x86_64.tar.gz'
Invoke-RestMethod -Uri $SRC -OutFile $DST
```

</details>

## ⚖️ License

[Apache-2.0](LICENSE)
