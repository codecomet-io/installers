package main

import (
	"fmt"
	"github.com/codecomet-io/go-sdk/codecomet"
	"github.com/codecomet-io/go-sdk/controller"
	"github.com/codecomet-io/go-sdk/execcontext/debian"
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

	// Get the official image here, so, custom resolver for Debian
	deb := debian.Debian{
		Version:  debian.Bullseye,
		Platform: plt,

		Resolver: func(img *codecomet.Image, version debian.Version) {
			img.Registry = "docker.io"
			img.Owner = "library"
			img.Name = "debian"
			img.Tag = string(version) + "-slim"
		},
	}

	// add eatmydata to speed things up and socat + nano to ease debugging out (XXX should not be in runtime images though...)
	deb.Apt().Install("eatmydata", "socat", "nano")

	outx := deb.GetInternalState()

	controller.Get().Exporter = &controller.Export{
		// Oci: "oci-tester/exp.tar",
		Images: []string{
			fmt.Sprintf("%s:%s-%s", pushToBase, debianVersion, plt.Architecture+plt.Variant),
		},
	}

	controller.Get().Do(outx)
}
