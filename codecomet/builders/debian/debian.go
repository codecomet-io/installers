package main

import (
	"fmt"
	"github.com/codecomet-io/go-sdk/base/debian"
	"github.com/codecomet-io/go-sdk/bin/apt"
	"github.com/codecomet-io/go-sdk/codecomet"
	"github.com/codecomet-io/go-sdk/controller"
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
	supportedPlatforms = []*codecomet.Platform{
		codecomet.LinuxAmd64,
		codecomet.LinuxArm64,
		codecomet.LinuxArmV7,
	}
)

func main() {
	for _, v := range supportedDebianVersions {
		for _, plt := range supportedPlatforms {
			build(v, plt)
		}
	}
}

func build(debianVersion debian.Version, plt *codecomet.Platform) {
	controller.Init()

	deb := (&codecomet.Image{
		Registry: "docker.io",
		Owner:    "library",
		Name:     "debian",
		Tag:      string(debianVersion) + "-slim",
		Platform: plt,
	})

	ap := apt.New(deb.GetInternalState())
	// add eatmydata to speed things up and socat to ease debugging out (XXX should not be in runtime images though...)
	ap.Install("eatmydata", "socat", "nano")

	outx := ap.State

	controller.Get().Exporter = &controller.Export{
		// Oci: "oci-tester/exp.tar",
		Images: []string{
			fmt.Sprintf("%s:%s-%s", pushToBase, debianVersion, plt.Architecture+plt.Variant),
		},
	}

	controller.Get().Do(outx)
}
