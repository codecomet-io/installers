package base

import (
	"fmt"
	"github.com/codecomet-io/installers/sdk/base/debian"
	"github.com/codecomet-io/installers/sdk/base/golang"
	"github.com/codecomet-io/installers/sdk/base/llvm"
	"github.com/codecomet-io/isovaline/sdk/codecomet"
	"github.com/codecomet-io/isovaline/sdk/wrapllb/platform"
	"github.com/moby/buildkit/client/llb"
)

const (
	registry   = "docker.io"
	ns         = "codecometio"
	debianName = "distro_debian"
	cName      = "builder_c"
	goName     = "builder_golang"
	pyName     = "builder_python"
)

func Debian(version debian.Version, plt *platform.Platform) llb.State {
	return codecomet.From(&codecomet.Image{
		Registry: registry,
		Owner:    ns,
		Name:     debianName,
		Tag:      fmt.Sprintf("%s-%s", version, plt.Architecture+plt.Variant),
		Platform: plt,
	})
}

func C(debianVersion debian.Version, llvmVersion llvm.Version, withMacOS bool, plt *platform.Platform) llb.State {
	withMac := ""
	if withMacOS {
		withMac = "-macos"
	}

	return codecomet.From(&codecomet.Image{
		Registry: registry,
		Owner:    ns,
		Name:     cName,
		Tag:      fmt.Sprintf("%s-%d%s-%s", debianVersion, llvmVersion, withMac, plt.Architecture+plt.Variant),
		Platform: plt,
	})
}

func Go(debianVersion debian.Version, golangVersion golang.Version, withCGO bool, withMacOS bool, plt *platform.Platform) llb.State {
	cgo := ""
	mac := ""
	if withCGO {
		cgo = "-cgo"
	}
	if withMacOS {
		mac = "-macos"
	}
	tag := fmt.Sprintf("%s-%s%s%s-%s", debianVersion, golangVersion, cgo, mac, plt.Architecture+plt.Variant)
	return codecomet.From(&codecomet.Image{
		Registry: registry,
		Owner:    ns,
		Name:     goName,
		Tag:      tag,
		Platform: plt,
	})
}

func Python(debianVersion debian.Version, withC bool, plt *platform.Platform) llb.State {
	c := ""
	if withC {
		c = "-c"
	}
	tag := fmt.Sprintf("%s%s-%s", debianVersion, c, plt.Architecture+plt.Variant)
	return codecomet.From(&codecomet.Image{
		Registry: registry,
		Owner:    ns,
		Name:     pyName,
		Tag:      tag,
		Platform: plt,
	})
}

/*
func Node(debianVersion debian.Version, withC bool, plt *platform.Platform) llb.State {
	c := ""
	if withC {
		c = "-c"
	}
	tag := fmt.Sprintf("%s%s-%s", debianVersion, c, plt.Architecture+plt.Variant)
	return codecomet.From(&codecomet.Image{
		Registry: registry,
		Owner:    ns,
		Name:     pyName,
		Tag:      tag,
		Platform: plt,
	})
}

*/
