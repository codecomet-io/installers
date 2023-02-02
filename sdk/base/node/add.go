package node

import (
	"fmt"
	"github.com/codecomet-io/installers/sdk/base"
	"github.com/codecomet-io/installers/sdk/base/debian"
	"github.com/codecomet-io/installers/sdk/bin/bash"
	"github.com/codecomet-io/isovaline/isovaline/core/log"
	"github.com/codecomet-io/isovaline/sdk/codecomet"
	"github.com/codecomet-io/isovaline/sdk/wrapllb"
	"github.com/codecomet-io/isovaline/sdk/wrapllb/platform"
	"github.com/moby/buildkit/client/llb"
	"github.com/opencontainers/go-digest"
	"net/url"
)

func getDownloadURL(version Version, plt *platform.Platform) *url.URL {
	var arch string
	// := plt.OS + "-" + plt.Architecture
	if plt.OS != platform.Linux {
		log.Fatal().Msgf("Unsupported platform %s %s %s", plt.OS, plt.Architecture, plt.Variant)
	}
	if plt.Architecture == platform.Arm {
		if plt.Variant == platform.V7 {
			arch = plt.OS + "-armv7l"
		} else if plt.Variant == platform.V6 {
			log.Fatal().Msgf("Unsupported platform %s %s %s", plt.OS, plt.Architecture, plt.Variant)
		}
	} else if plt.Architecture == platform.Arm64 {
		arch = plt.OS + "-arm64"
	} else if plt.Architecture == platform.Amd64 {
		arch = plt.OS + "-x64"
	} else {
		log.Fatal().Msgf("Unsupported platform %s %s %s", plt.OS, plt.Architecture, plt.Variant)
	}
	// https://nodejs.org/dist/v18.13.0/node-v18.13.0-linux-armv7l.tar.xz
	// https://nodejs.org/dist/v18.13.0/node-v18.13.0-linux-arm64.tar.xz
	// https://nodejs.org/dist/v18.13.0/node-v18.13.0-linux-x64.tar.xz
	return &url.URL{
		Scheme: "https",
		Host:   "nodejs.org",
		Path:   fmt.Sprintf("/dist/%s/node-%s-%s.tar.xz", version, version, arch),
	}
}

func Add(nodeVersion Version, nodeChecksum digest.Digest, plt *platform.Platform) llb.State {
	grp := &wrapllb.Group{
		Name: "node",
		ID:   fmt.Sprintf("Installing node %s for %s", nodeVersion, plt.Architecture+plt.Variant),
	}

	fromHttp := codecomet.From(&codecomet.HTTP{
		URL:    getDownloadURL(nodeVersion, plt),
		Digest: nodeChecksum,
		Output: "node.tar.xz",
		Base: codecomet.Base{
			Group: grp,
		},
	})

	ll := []llb.ConstraintsOpt{
		llb.ProgressGroup(grp.ID, grp.Name, grp.DoNotDisplay),
	}

	st := wrapllb.Copy(
		fromHttp,
		"/node.tar.xz",
		llb.Scratch().File(llb.Mkdir("/opt", 0755, llb.WithParents(true)), ll...),
		"/opt",
		&wrapllb.CopyOptions{
			Unpack: true,
		},
	)

	bsh := bash.New(base.Debian(debian.Bullseye, platform.DefaultPlatform))
	bsh.ReadOnly = true
	bsh.Mount["/opt"] = &wrapllb.State{
		Source: st,
		Path:   "/opt",
	}

	bsh.Run("Symlinking", `
		cd /opt
		ln -s node-* node
	`)

	return bsh.Mount["/opt"].Source
}
