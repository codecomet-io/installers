package main

import (
	"fmt"
	"github.com/codecomet-io/go-sdk/base"
	"github.com/codecomet-io/go-sdk/base/debian"
	"github.com/codecomet-io/go-sdk/base/llvm"
	"github.com/codecomet-io/go-sdk/base/node"
	"github.com/codecomet-io/go-sdk/codecomet"
	"github.com/codecomet-io/go-sdk/controller"
	"github.com/codecomet-io/go-sdk/coretypes"
	"github.com/moby/buildkit/client/llb"
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
			Platform:     coretypes.LinuxArm64,
			NodeChecksum: node.Node19DigestArm64,
		},
		{
			NodeVersion:  node.Node19,
			Platform:     coretypes.LinuxAmd64,
			NodeChecksum: node.Node19DigestAmd64,
		},
		{
			NodeVersion:  node.Node19,
			Platform:     coretypes.LinuxArmV7,
			NodeChecksum: node.Node19DigestArm7,
		},
		{
			NodeVersion:  node.Node18,
			Platform:     coretypes.LinuxArm64,
			NodeChecksum: node.Node18DigestArm64,
		},
		{
			NodeVersion:  node.Node18,
			Platform:     coretypes.LinuxAmd64,
			NodeChecksum: node.Node18DigestAmd64,
		},
		{
			NodeVersion:  node.Node18,
			Platform:     coretypes.LinuxArmV7,
			NodeChecksum: node.Node18DigestArm7,
		},
	}
)

type Target struct {
	NodeVersion  node.Version
	NodeChecksum coretypes.Digest
	Platform     *coretypes.Platform
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
				if target.Platform != coretypes.LinuxArmV7 {
					build(target.NodeVersion, target.NodeChecksum, debVersion, llvmVersion, true, target.Platform)
				}
			}
		}
	}
}

func build(nodeVersion node.Version, nodeChecksum coretypes.Digest, debianVersion debian.Version, llvmVersion llvm.Version, withC bool, plt *coretypes.Platform) {
	codecomet.Init()

	c := ""

	var bb llb.State
	bb = base.Debian(debianVersion, plt)
	if withC {
		c = "-c"
		bb = base.C(debianVersion, llvmVersion, false, plt)
	}

	outx := llb.Merge([]llb.State{bb, node.Add(nodeVersion, nodeChecksum, plt)})

	tag := fmt.Sprintf("%s-%s%s-%s", debianVersion, nodeVersion, c, plt.Architecture+plt.Variant)

	controller.Get().Exporter = &controller.Export{
		Images: []string{
			"docker.io/codecometio/builder_node:" + tag,
		},
	}

	controller.Get().Do(outx)
}
