package main

import (
	"fmt"
	"github.com/codecomet-io/go-sdk/base"
	"github.com/codecomet-io/go-sdk/base/debian"
	"github.com/codecomet-io/go-sdk/base/llvm"
	"github.com/codecomet-io/go-sdk/bin/apt"
	"github.com/codecomet-io/go-sdk/codecomet"
	"github.com/codecomet-io/go-sdk/controller"
	"github.com/codecomet-io/go-sdk/coretypes"
	"github.com/moby/buildkit/client/llb"
)

var (
	supportedLLVM = []llvm.Version{
		llvm.V15,
	}

	supportedDebianVersions = []debian.Version{
		debian.Bullseye,
		debian.Sid,
	}
	supportedPlatforms = []*coretypes.Platform{
		coretypes.LinuxArm64,
		coretypes.LinuxAmd64,
		coretypes.LinuxArmV7,
	}
)

func main() {
	for _, llvmVersion := range supportedLLVM {
		for _, debVersion := range supportedDebianVersions {
			for _, target := range supportedPlatforms {
				build(debVersion, llvmVersion, false, target)
				// No c variant on armhf
				if target != coretypes.LinuxArmV7 {
					build(debVersion, llvmVersion, true, target)
				}
			}
		}
	}
}

func build(debianVersion debian.Version, llvmVersion llvm.Version, withC bool, plt *coretypes.Platform) {
	codecomet.Init()

	c := ""

	var bb llb.State
	bb = base.Debian(debianVersion, plt)
	if withC {
		c = "-c"
		bb = base.C(debianVersion, llvmVersion, false, plt)
	}

	ag := apt.New(bb)
	ag.Install("python3", "python3-pip", "python3-venv")
	outx := ag.State

	tag := fmt.Sprintf("%s%s-%s", debianVersion, c, plt.Architecture+plt.Variant)

	controller.Get().Exporter = &controller.Export{
		Images: []string{
			"docker.io/codecometio/builder_python:" + tag,
		},
	}

	controller.Get().Do(outx)
}
