# CodeComet installers

## Installation

### Requirements

This is for macOS only for now (tested on Ventura).

```shell
# Make sure homebrew is installed
# See https://brew.sh/ for instructions
# Usually, this is:
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

### Do

Using the tap:

```shell
brew tap codecomet-io/tap https://github.com/codecomet-io/installers.git
brew install --HEAD codecomet
# Then start the service
codecomet-machine install
```

Give it some time to boot, then you can verify it is up with:

```shell
codecomet-machine status
```

### Caveats

Currently, the formula conflicts with `lima`.
The workaround unfortunately is to `brew unlink lima`.

Lima VZ is not usable rn: https://github.com/lima-vm/lima/issues/1200

Lima clock appears to drift is some unclear circumstances: https://github.com/lima-vm/lima/issues/1307#issuecomment-1397996400

