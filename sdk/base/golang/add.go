package golang

import (
	"fmt"
	"github.com/codecomet-io/isovaline/isovaline/core/log"
	"github.com/codecomet-io/isovaline/sdk/codecomet"
	"github.com/codecomet-io/isovaline/sdk/wrapllb"
	"github.com/codecomet-io/isovaline/sdk/wrapllb/platform"
	"github.com/moby/buildkit/client/llb"
	"github.com/opencontainers/go-digest"
	"net/url"
)

func getGoDownloadURL(version Version, plt *platform.Platform) *url.URL {
	arch := plt.OS + "-" + plt.Architecture
	if plt.OS != platform.Linux {
		log.Fatal().Msgf("Unsupported platform %s %s %s", plt.OS, plt.Architecture, plt.Variant)
	}
	if plt.Architecture == platform.Arm {
		arch = plt.OS + "-armv6l"
	}

	return &url.URL{
		Scheme: "https",
		Host:   "dl.google.com",
		Path:   fmt.Sprintf("/go/go%s.%s.tar.gz", version, arch),
	}
}

func Add(goVersion Version, goChecksum digest.Digest, plt *platform.Platform) llb.State {
	grp := &wrapllb.Group{
		Name: "golang",
		ID:   fmt.Sprintf("Installing golang %s for %s", goVersion, plt.Architecture+plt.Variant),
	}
	fromHttp := codecomet.From(&codecomet.HTTP{
		URL:    getGoDownloadURL(goVersion, plt),
		Digest: goChecksum,
		Output: "go.tar.gz",
		Base: codecomet.Base{
			Group: grp,
		},
	})

	ll := []llb.ConstraintsOpt{
		llb.ProgressGroup(grp.ID, grp.Name, grp.DoNotDisplay),
	}

	return wrapllb.Copy(
		fromHttp,
		"/go.tar.gz",
		llb.Scratch().File(llb.Mkdir("/opt", 0755, llb.WithParents(true)), ll...),
		"/opt",
		&wrapllb.CopyOptions{
			Unpack: true,
		},
	)
}
