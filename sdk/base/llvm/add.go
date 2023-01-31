package llvm

import (
	"fmt"
	"github.com/codecomet-io/installers/sdk/base/debian"
	"github.com/codecomet-io/installers/sdk/debian/apt"
	"github.com/codecomet-io/isovaline/isovaline/core/log"
	"github.com/codecomet-io/isovaline/sdk/wrapllb"
	"github.com/moby/buildkit/client/llb"
	"net/url"
)

func Add(debianState llb.State, debianVersion debian.Version, llvmVersion Version) llb.State {
	if debianVersion == debian.Bookworm {
		log.Fatal().Msg("LLVM is not available for Debian Bookworm")
	}

	// Get an apt
	aptGet := apt.New(debianState)
	aptGet.Group = &wrapllb.Group{
		ID:           fmt.Sprintf("LLVM-%d-%s", llvmVersion, debianVersion),
		Name:         fmt.Sprintf("Installing LLVM %d on Debian %s ", llvmVersion, debianVersion),
		DoNotDisplay: false,
	}

	// Get ca certs so we can add llvm.org
	// XXX should be a conditional
	aptGet.Install("ca-certificates")

	// Packages for llvm
	packages := []interface{}{
		fmt.Sprintf("llvm-%d", llvmVersion),
		fmt.Sprintf("clang-%d", llvmVersion),
		fmt.Sprintf("lld-%d", llvmVersion),
		fmt.Sprintf("llvm-%d-dev", llvmVersion),
	}

	// Prep up the deb repo
	pc := fmt.Sprintf("llvm-toolchain-%s", debianVersion)
	if debianVersion == debian.Sid {
		pc = "llvm-toolchain"
		debianVersion = debian.Unstable
	}
	if llvmVersion != 17 {
		pc = fmt.Sprintf("%s-%d", pc, llvmVersion)
	}
	src := fmt.Sprintf("deb https://apt.llvm.org/%s/ %s main", debianVersion, pc)

	// Add it
	aptGet.AddList("llvm", src, &url.URL{
		Scheme: "https",
		Host:   "apt.llvm.org",
		Path:   "/llvm-snapshot.gpg.key",
	})

	// Go install all of that
	aptGet.Install(packages...)

	return aptGet.State
}
