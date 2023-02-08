package main

import (
	"fmt"
	"github.com/codecomet-io/go-sdk/base"
	"github.com/codecomet-io/go-sdk/base/c"
	"github.com/codecomet-io/go-sdk/base/debian"
	"github.com/codecomet-io/go-sdk/base/llvm"
	"github.com/codecomet-io/go-sdk/base/macos"
	"github.com/codecomet-io/go-sdk/bin/bash"
	"github.com/codecomet-io/go-sdk/codecomet"
	"github.com/codecomet-io/go-sdk/controller"
	"github.com/codecomet-io/go-sdk/coretypes"
	"github.com/moby/buildkit/client/llb"
)

const (
	pushToBase = "docker.io/codecometio/builder_c"
)

var (
	// Variants debian we build
	supportedDebianVersions = []debian.Version{
		debian.Bullseye,
		// debian.Sid,
		// llvm org packages are not available for bookworm
		// debian.Bookworm,
		// Buster is long dead
		// debian.Buster,
	}
	// Platforms we build
	supportedPlatforms = []*coretypes.Platform{
		coretypes.LinuxArm64,
		coretypes.LinuxAmd64,
	}
	// Versions of LLVM we build
	supportedLLVM = []llvm.Version{
		llvm.V15,
		// llvm.V17,
		// llvm.V16,
	}

	// Platforms we can cross compile to
	supportedTargetPlatforms = []*coretypes.Platform{
		coretypes.LinuxAmd64,
		coretypes.LinuxArm64,
		coretypes.LinuxArmV7,
	}
)

func main() {
	for _, plt := range supportedPlatforms {
		for _, llvmVersion := range supportedLLVM {
			for _, v := range supportedDebianVersions {
				build(v, llvmVersion, true, plt)
			}
		}
	}
	for _, plt := range supportedPlatforms {
		for _, llvmVersion := range supportedLLVM {
			for _, v := range supportedDebianVersions {
				build(v, llvmVersion, false, plt)
			}
		}
	}
}

func build(debianVersion debian.Version, llvmVersion llvm.Version, withMacOS bool, plt *coretypes.Platform) {
	codecomet.Init()

	// Get the requested Debian
	state := base.Debian(debianVersion, plt)

	// Add basic C stuff
	state = c.Add(state, supportedTargetPlatforms)

	// Now add LLVM
	state = llvm.Add(state, debianVersion, llvmVersion)

	// Need this... Should be in the llvm overlay actually
	bsh := bash.New(state)
	bsh.Run("Symlinking", fmt.Sprintf(`
		# Symlink clang - not sure why the package does not do that on its own
		ln -s /usr/lib/llvm-%d/bin/clang /usr/bin/clang
		ln -s /usr/lib/llvm-%d/bin/clang++ /usr/bin/clang++

		# Dirty trick to carry in compiler-rt in case it gets compiled
		ln -s /opt/macosxcross/compiler-rt/lib /usr/lib/llvm-%d/lib/clang/%d*/
		ln -s /opt/macosxcross/compiler-rt/include/sanitizer /usr/lib/llvm-%d/lib/clang/%d*/include/
	`, llvmVersion, llvmVersion, llvmVersion, llvmVersion, llvmVersion, llvmVersion))

	// Optionally add the macOS toolchain
	withMac := ""
	if withMacOS {
		// Now, we can pack
		mt := macos.Add("/Applications/Xcode.app", macos.V13_1, debianVersion, llvmVersion, plt)
		// And merge the toolkit in
		state = llb.Merge([]llb.State{bsh.State, mt})
		withMac = "-macos"
	}

	controller.Get().Exporter = &controller.Export{
		Images: []string{
			fmt.Sprintf("%s:%s-%d%s-%s", pushToBase, debianVersion, llvmVersion, withMac, plt.Architecture+plt.Variant),
		},
		// Oci: "tmp/test-out-oci.tar.gz",
		// Tarball: "tmp/test-out.tar",
		// Local: os.Args[1],
	}

	controller.Get().Do(state)

}
