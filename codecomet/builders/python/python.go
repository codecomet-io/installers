package main

import (
	"fmt"
	"github.com/codecomet-io/go-sdk/codecomet"
	"github.com/codecomet-io/go-sdk/controller"
	"github.com/codecomet-io/go-sdk/execcontext/debian"
	"github.com/codecomet-io/go-sdk/overlay/c"
	"github.com/codecomet-io/go-sdk/overlay/llvm"
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

	ag := deb.Apt()
	ag.Install("python3", "python3-pip", "python3-venv")

	controller.Get().Exporter = &controller.Export{
		Images: []string{
			"docker.io/codecometio/builder_python:" + fmt.Sprintf("%s%s-%s", debianVersion, cTag, plt.Architecture+plt.Variant),
		},
	}

	controller.Get().Do(deb.GetInternalState())
}
