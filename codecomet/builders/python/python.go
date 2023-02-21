package main

import (
	"fmt"
	"github.com/codecomet-io/go-sdk/base/debian"
	"github.com/codecomet-io/go-sdk/base/llvm"
	"github.com/codecomet-io/go-sdk/bin/apt"
	"github.com/codecomet-io/go-sdk/codecomet"
	"github.com/codecomet-io/go-sdk/controller"
	"github.com/codecomet-io/go-sdk/root"
)

var (
	supportedLLVM = []llvm.Version{
		llvm.V15,
	}

	supportedDebianVersions = []debian.Version{
		debian.Bullseye,
		debian.Sid,
	}
	supportedPlatforms = []*codecomet.Platform{
		codecomet.LinuxArm64,
		codecomet.LinuxAmd64,
		codecomet.LinuxArmV7,
	}
)

func main() {
	for _, llvmVersion := range supportedLLVM {
		for _, debVersion := range supportedDebianVersions {
			for _, target := range supportedPlatforms {
				build(debVersion, llvmVersion, false, target)
				// No c variant on armhf
				if target != codecomet.LinuxArmV7 {
					build(debVersion, llvmVersion, true, target)
				}
			}
		}
	}
}

func build(debianVersion debian.Version, llvmVersion llvm.Version, withC bool, plt *codecomet.Platform) {
	controller.Init()

	c := ""
	bb := root.Debian(debianVersion, plt)
	if withC {
		c = "-c"
		bb = root.C(debianVersion, llvmVersion, false, plt)
	}

	ag := apt.New(bb.GetInternalState())
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
