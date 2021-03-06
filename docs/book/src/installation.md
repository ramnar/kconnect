# Installation

- [Mac OSX](#mac)
- [Linux](#linux)
- [Windows](#windows)
- [Docker](#docker)

<em>NOTE:</em> `kconnect` requires [aws-iam-authenticator](https://github.com/kubernetes-sigs/aws-iam-authenticator) to authenticate to the AWS EKS cluster provider.

## Mac

To install on OSX you can use homebrew:

```bash
brew install fidelity/tap/kconnect
```

Alternatively you can download a binary from the latest [release](https://github.com/fidelity/kconnect/releases).

## Linux

The latest [release](https://github.com/fidelity/kconnect/releases) contains **.deb**, **.rpm** and binaries for Linux.

We are working on publishing as a snap.

## Windows

The latest [release](https://github.com/fidelity/kconnect/releases) contains a binary for Windows.

We have an open issue to support chocolatey in the future/

## Docker

You can also use kconnect via Docker by using the images we publish to Docker Hub:

```bash
docker pull docker.io/kconnectcli/kconnect:latest
docker run -it --rm -v ~/.kconnect:/.kconnect kconnect:latest use eks --idp-protocol saml
```
