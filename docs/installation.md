# Installation

You have the following options to install the operator-builder CLI:
* [Download the latest binary with your browser](https://github.com/nukleros/operator-builder/releases/latest)
* [Download with wget](#wget)
* [Homebrew](#homebrew)
* [Snap](#snap)
* [Docker Image](#docker-image)
* [Go Install](#go-install)

### wget
Use wget to download the pre-compiled binaries:

```bash
wget https://github.com/vmware-tanzu-labs/operator-builder/releases/download/${VERSION}/${BINARY}.tar.gz -O - |\
  tar xz && sudo mv operator-builder /usr/bin/operator-builder
```

For instance, VERSION=v0.5.0 and BINARY=operator-builder_${VERSION}_Linux_x86_64

### Homebrew

Available for Mac and Linux.

Using [Homebrew](https://brew.sh/)  

```bash
brew tap vmware-tanzu-labs/tap
brew install operator-builder
```

### Snap

Available for Linux only.

```bash
snap install operator-builder
```

>**NOTE**: `operator-builder` installs with [_strict confinement_](https://docs.snapcraft.io/snap-confinement/6233) in snap, this means it doesn't have direct access to root files.

### Docker Image

```bash
docker pull ghcr.io/vmawre-tanzu-labs/operator-builder
```

#### One-shot container use

```bash
docker run --rm -v "${PWD}":/workdir ghcr.io/vmware-tanzu-labs/operator-builder [flags]
```


#### Run container commands interactively

```bash
docker run --rm -it -v "${PWD}":/workdir --entrypoint sh ghcr.io/vmawre-tanzu-labs/operator-builder
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

