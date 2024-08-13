# Installation

You have the following options to install the operator-builder CLI:

- [Download the latest binary with your browser](https://github.com/nukleros/operator-builder/releases/latest)
- [Download with wget](#wget)
- [Homebrew](#homebrew)
- [Docker Image](#docker-image)
- [Go Install](#go-install)

### wget
Use wget to download the pre-compiled binaries:

```bash
VERSION=v0.11.0
OS=Linux
ARCH=x86_64
wget https://github.com/nukleros/operator-builder/releases/download/${VERSION}/operator-builder_${VERSION}_${OS}_${ARCH}.tar.gz -O - |\
    tar -xz && sudo mv operator-builder /usr/local/bin/operator-builder
```

### Homebrew

Available for Mac and Linux.

Using [Homebrew](https://brew.sh/)

```bash
brew tap nukleros/tap
brew install nukleros/tap/operator-builder
```

### Docker Image

```bash
docker pull ghcr.io/nukleros/operator-builder
```

#### One-shot container use

```bash
docker run --rm -v "${PWD}":/workdir ghcr.io/nukleros/operator-builder [flags]
```

#### Run container commands interactively

```bash
docker run --rm -it -v "${PWD}":/workdir --entrypoint sh ghcr.io/nukleros/operator-builder
```

It can be useful to have a bash function to avoid typing the whole docker command:

```bash
operator-builder() {
  docker run --rm -i -v "${PWD}":/workdir ghcr.io/nukleros/operator-builder "$@"
}
```

### Go Install

```bash
go install github.com/nukleros/operator-builder/cmd/operator-builder@latest
```
