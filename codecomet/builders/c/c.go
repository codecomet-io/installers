package main

import (
	"fmt"
	"github.com/codecomet-io/go-sdk/codecomet"
	"github.com/codecomet-io/go-sdk/controller"
	"github.com/codecomet-io/go-sdk/execcontext/debian"
	"github.com/codecomet-io/go-sdk/fileset"
	"github.com/codecomet-io/go-sdk/overlay/c"
	"github.com/codecomet-io/go-sdk/overlay/llvm"
	"github.com/codecomet-io/go-sdk/overlay/macos"
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
	supportedPlatforms = []*codecomet.Platform{
		codecomet.LinuxArm64,
		codecomet.LinuxAmd64,
	}
	// Versions of LLVM we build
	supportedLLVM = []llvm.Version{
		llvm.V15,
		// llvm.V17,
		// llvm.V16,
	}

	// Platforms we can cross compile to
	supportedTargetPlatforms = codecomet.DefaultPlatformSet
	/*[]*codecomet.Platform{
		codecomet.LinuxAmd64,
		codecomet.LinuxArm64,
		codecomet.LinuxArmV7,
	}*/

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

func build(debianVersion debian.Version, llvmVersion llvm.Version, withMacOS bool, plt *codecomet.Platform) {
	controller.Init()

	deb := &debian.Debian{
		Version:  debianVersion,
		Platform: plt,
	}

	// XXX should Overlay be a generic action mechanism that alters the input?
	// eg: fileset.Overlay(func(codecomet.Fileset))

	c.Overlay(deb, supportedTargetPlatforms)
	llvm.Overlay(deb, llvmVersion)

	// Need this... Should be in the llvm overlay actually
	/*
		bsh := bash.New(deb.GetInternalState())
		bsh.Use(deb)
		bsh.RunOLD("Symlinking", fmt.Sprintf(`
			# Symlink clang - not sure why the package does not do that on its own
			ln -s /usr/lib/llvm-%d/bin/clang /usr/bin/clang
			ln -s /usr/lib/llvm-%d/bin/clang++ /usr/bin/clang++

			# Dirty trick to carry in compiler-rt in case it gets compiled
			ln -s /opt/macosxcross/compiler-rt/lib /usr/lib/llvm-%d/lib/clang/%d*     /
			ln -s /opt/macosxcross/compiler-rt/include/sanitizer /usr/lib/llvm-%d/lib/clang/%d*    /include/
		`, llvmVersion, llvmVersion, llvmVersion, llvmVersion, llvmVersion, llvmVersion))
	*/

	// Optionally add the macOS toolchain
	withMac := ""
	if withMacOS {
		// Now, we can pack
		mt := macos.Add("/Applications/Xcode.app", macos.V13_1, debianVersion, llvmVersion, plt)
		// And merge the toolkit in
		deb.Adopt(fileset.Merge([]codecomet.FileSet{deb, mt}, &codecomet.MergeOptions{}).GetInternalState())
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

	controller.Get().Do(deb.GetInternalState())
}
