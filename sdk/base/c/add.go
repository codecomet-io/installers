package c

import (
	"fmt"
	"github.com/codecomet-io/installers/sdk/debian/apt"
	"github.com/codecomet-io/isovaline/sdk/wrapllb"
	"github.com/codecomet-io/isovaline/sdk/wrapllb/platform"
	"github.com/moby/buildkit/client/llb"
	"strings"
)

var (
	// Packages we install for the current architecture
	basePackages = []string{
		"make",
		"cmake",
		"automake", // implies autoconf
		"pkg-config",
	}
	// Packages we install for all architectures
	crossPackages = []string{
		"musl",
		"musl-dev",
		"libc6",
		"libc6-dev",
		// "gcc",
		// "g++",
		// "libtool",
		// "dpkg-dev",
		// "dpkg-cross",
	}
)

func Add(debianState llb.State, supportedTargetPlatforms []*platform.Platform, additionalMultiPackages ...string) llb.State {
	// Get an apt
	aptGet := apt.New(debianState)
	se := []string{}
	for _, v := range supportedTargetPlatforms {
		se = append(se, v.Architecture+v.Variant)
	}
	aptGet.Group = &wrapllb.Group{
		ID:           fmt.Sprintf("C-%s", strings.Join(se, " ")),
		Name:         fmt.Sprintf("Installing C/C++ crossbuild essentials for %s", strings.Join(se, " ")),
		DoNotDisplay: false,
	}

	// The list of essential build packages
	aptGet.AddArchitecture(supportedTargetPlatforms)
	packages := []interface{}{}
	for _, pack := range basePackages {
		packages = append(packages, pack)
	}
	for _, plat := range supportedTargetPlatforms {
		// Better than installing piecemeal, as there are some conflicts trying to get g**:arch
		packages = append(packages, fmt.Sprintf("crossbuild-essential-%s", apt.PlatformToDebianArchitecture(plat)))
		for _, pack := range crossPackages {
			packages = append(packages, &apt.Package{
				Name:     pack,
				Platform: plat,
			})
		}
		for _, pack := range additionalMultiPackages {
			packages = append(packages, &apt.Package{
				Name:     pack,
				Platform: plat,
			})
		}
	}

	// Go install all of that
	aptGet.Install(packages...)

	return aptGet.State
}
