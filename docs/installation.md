# Installation

You have the following options to install the operator-builder CLI:
* [Download the latest binary with your browser](https://github.com/nukleros/operator-builder/releases/latest)
* [Download with wget](#wget)
* [Homebrew](#homebrew)
* [Docker Image](#docker-image)
* [Go Install](#go-install)
* [Snap](#snap)

### wget
Use wget to download the pre-compiled binaries:

```bash
VERSION=v0.6.0
OS=Linux
ARCH=x86_64
wget https://github.com/nukleros/operator-builder/releases/download/${VERSION}/operator-builder_${VERSION}_${OS}_${ARCH}.gz -O - |\
    gzip -d && sudo mv operator-builder_${VERSION}_${OS}_${ARCH} /usr/local/bin/operator-builder
```

### Homebrew

Available for Mac and Linux.

Using [Homebrew](https://brew.sh/)

```bash
brew tap vmware-tanzu-labs/tap
brew install operator-builder
```

### Docker Image

```bash
docker pull ghcr.io/nukleros/operator-builder
```

#### One-shot container use

```bash
docker run --rm -v "${PWD}":/workdir ghcr.io/vmware-tanzu-labs/operator-builder [flags]
```

#### Run container commands interactively

```bash
docker run --rm -it -v "${PWD}":/workdir --entrypoint sh ghcr.io/nukleros/operator-builder
```

It can be useful to have a bash function to avoid typing the whole docker command:

```bash
operator-builder() {
  docker run --rm -i -v "${PWD}":/workdir ghcr.io/vmware-tanzu-labs/operator-builder "$@"
}
```

### Go Install

```bash
GO111MODULE=on go get github.com/vmware-tanzu-labs/operator-builder/cmd/operator-builder
```

### Snap

**NOTE:** support for Snaps has been removed due to, what we feel, is increasingly unstable developer experience in publishing snaps.  We can readdress 
this at such a time where stability to the project has returned.  As of now, the latest available snap for operator-builder is v0.5.0.

Available for Linux only.

```bash
snap install operator-builder
```

>**NOTE**: `operator-builder` installs with [_strict confinement_](https://docs.snapcraft.io/snap-confinement/6233) in snap, this means it doesn't have direct access to root files.
