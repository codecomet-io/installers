# CodeComet installers

## Install on macOS

### Requirements

```shell
# Make sure homebrew is installed
# See https://brew.sh/ for instructions
# Usually, this is:
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
```

### Installation

Using the tap:

```shell
brew tap codecomet-io/tap https://github.com/codecomet-io/installers.git
brew install --HEAD codecomet
codecomet-machine install
```

### Caveats

Currently, the formula conflicts with `lima`.
The workaround unfortunately is to `brew unlink lima`.

<!--
Debugging the tap locally:
```shell
brew install --HEAD --build-from-source ./codecomet.rb
```
#  --only-dependencies
-->

<!--
Or alternatively the install script:

```shell
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/codecomet-io/installers/HEAD/install.sh)"
```
-->



<!--

## Hacking on the tap


## [WORK IN PROGRESS] On-boarding macOS users into the private beta

1. **Optional**: for users who do NOT have an existing ssh key or want to use a new one, have
them create one, for example something like:

```shell
ssh-keygen -t ed25519 -C "john@private.codecomet.io" -f ~/.ssh/beta.codecomet.io
```

2. Have the user edit `~/.ssh/config` to add these lines:
```yaml
Host beta
    # Alternatively, point the identity file to whatever key they want to use
    identityfile ~/.ssh/beta.codecomet.io
    identitiesonly yes
    hostname github.com
    port 22
    user git
```

3. Have the user send us the public part of their key
```shell
# For example (if created above):
cat ~/.ssh/beta.codecomet.io.pub
# Or
# cat ~/.ssh/id_ed25519.pub
# Or whatever is the path to their public id
```

4. In our `installers` repository:
  * in `Settings > Deploy Keys > Add New`
  * do NOT give away write access
  * copy their public key from above in there

5. Repeat previous step for the `commander` repository

6. Have them clone this repository and run the install script
```shell
git clone git@github.com:codecomet-io/installers.git
cd installers
./install.sh
```

## Distribution graph

 * direct dependencies (QEMU & Go) are not pinned, and formulas are retrieved from brew central
 * indirect dependencies are also out there

## TODO

MVP:
 * \o?

Next:
 * check on this for private homebrew / stuff:
   * https://medium.com/prodopsio/creating-homebrew-taps-for-private-internal-tools-c41363d58ab0
   * https://gist.github.com/mlafeldt/8e7d50ee0b1de44e256d
   * https://franzramadhan.com/posts/9-publish-homebrew-in-private-repo/
 * have a complete installer spinning a private brew (https://github.com/dubo-dubon-duponey/tarmac/blob/master/init )
 * look at codesigner: https://dantheman827.github.io/ios-app-signer/

-->