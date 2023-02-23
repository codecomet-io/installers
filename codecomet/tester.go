package main

import (
	"fmt"
	"github.com/codecomet-io/go-sdk/codecomet"
	"github.com/codecomet-io/go-sdk/codecomet/core"
	"github.com/codecomet-io/go-sdk/controller"
	"github.com/codecomet-io/go-sdk/execcontext/debian"
	"github.com/codecomet-io/go-sdk/overlay/c"
	"github.com/codecomet-io/go-sdk/overlay/llvm"
	"github.com/codecomet-io/go-sdk/overlay/node"
)

var (
	// Variants debian we build
	supportedDebianVersions = []debian.Version{
		debian.Bullseye,
		debian.Sid,
	}

	// Versions of LLVM we build
	supportedLLVM = []llvm.Version{
		llvm.V15,
	}

	supportedTargets = []*Target{
		{
			NodeVersion:  node.Node19,
			Platform:     codecomet.LinuxArm64,
			NodeChecksum: node.Node19DigestArm64,
		},
		{
			NodeVersion:  node.Node19,
			Platform:     codecomet.LinuxAmd64,
			NodeChecksum: node.Node19DigestAmd64,
		},
		{
			NodeVersion:  node.Node19,
			Platform:     codecomet.LinuxArmV7,
			NodeChecksum: node.Node19DigestArm7,
		},
		{
			NodeVersion:  node.Node18,
			Platform:     codecomet.LinuxArm64,
			NodeChecksum: node.Node18DigestArm64,
		},
		{
			NodeVersion:  node.Node18,
			Platform:     codecomet.LinuxAmd64,
			NodeChecksum: node.Node18DigestAmd64,
		},
		{
			NodeVersion:  node.Node18,
			Platform:     codecomet.LinuxArmV7,
			NodeChecksum: node.Node18DigestArm7,
		},
	}
)

type Target struct {
	NodeVersion  node.Version
	NodeChecksum core.Digest
	Platform     *codecomet.Platform
}

/*
Notes:
- progress group, custom names, are too hard to reach
- mount state is counter-intuitive
- notion of unique identifier for an action is problematic - what truly uniquely define an action, that should be preserved over time?
	- if its the name, we have to make sure you cannot have a duplicate name
	- at least separate display/description/human from unique identifier of the action
- inability to specify FOO=something$FOO is a problem (specifically for PATH) - GetEnv?
- when pushing a resulting image, environment is NOT preserved - likely need to manipulate ImageConfig

Parameterizing plans is WRONG
Components therefore should NOT be plans.
*/

func main() {
	for _, llvmVersion := range supportedLLVM {
		for _, debVersion := range supportedDebianVersions {
			for _, target := range supportedTargets {
				build(target.NodeVersion, target.NodeChecksum, debVersion, llvmVersion, false, target.Platform)
				// No c variant on armhf
				if target.Platform != codecomet.LinuxArmV7 {
					build(target.NodeVersion, target.NodeChecksum, debVersion, llvmVersion, true, target.Platform)
				}
			}
		}
	}
}

func build(nodeVersion node.Version, nodeChecksum core.Digest, debianVersion debian.Version, llvmVersion llvm.Version, withC bool, plt *codecomet.Platform) {
	controller.Init()

	deb := &debian.Debian{
		Version:  debianVersion,
		Platform: plt,
	}

	// If we want C, use the c resolver
	cTag := ""
	if withC {
		cTag = "-c"
		// Alternatively, use the resolver to get the pre-built image?
		// deb.Resolver = overlay.WithC(llvmVersion, false)
		c.Overlay(deb, codecomet.DefaultPlatformSet)
		llvm.Overlay(deb, llvmVersion)
	}

	node.Overlay(deb, nodeVersion, nodeChecksum, plt)

	outx := deb.GetInternalState()

	controller.Get().Exporter = &controller.Export{
		Images: []string{
			"docker.io/codecometio/builder_node:" + fmt.Sprintf("%s-%s%s-%s", debianVersion, nodeVersion, cTag, plt.Architecture+plt.Variant),
		},
	}

	controller.Get().Do(outx)
}