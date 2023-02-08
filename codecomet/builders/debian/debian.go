package main

import (
	"fmt"
	"github.com/codecomet-io/go-sdk/base/debian"
	"github.com/codecomet-io/go-sdk/bin/apt"
	"github.com/codecomet-io/go-sdk/codecomet"
	"github.com/codecomet-io/go-sdk/controller"
	"github.com/codecomet-io/go-sdk/coretypes"
	"github.com/moby/buildkit/client/llb"
)

const (
	// preferedRegistry = "10.0.0.87:5000"
	pushToBase = "docker.io/codecometio/distro_debian"
)

var (
	supportedDebianVersions = []debian.Version{
		debian.Buster,
		debian.Bullseye,
		debian.Bookworm,
		debian.Sid,
	}
	supportedPlatforms = []*coretypes.Platform{
		coretypes.LinuxAmd64,
		coretypes.LinuxArm64,
		coretypes.LinuxArmV7,
	}
)

func main() {
	for _, v := range supportedDebianVersions {
		for _, plt := range supportedPlatforms {
			build(v, plt)
		}
	}
}

func build(debianVersion debian.Version, plt *coretypes.Platform) {
	codecomet.Init()

	outx := b(debianVersion, plt)

	controller.Get().Exporter = &controller.Export{
		// Oci: "oci-tester/exp.tar",
		Images: []string{
			fmt.Sprintf("%s:%s-%s", pushToBase, debianVersion, plt.Architecture+plt.Variant),
		},
	}

	controller.Get().Do(outx)
}

func b(debianVersion debian.Version, plt *coretypes.Platform) llb.State {
	deb := codecomet.From(&codecomet.Image{
		Registry: "docker.io",
		Owner:    "library",
		Name:     "debian",
		Tag:      string(debianVersion) + "-slim",
		Platform: plt,
	})

	ap := apt.New(deb)
	// add eatmydata to speed things up and socat to ease debugging out (XXX should not be in runtime images though...)
	ap.Install("eatmydata", "socat")

	return ap.State
}
