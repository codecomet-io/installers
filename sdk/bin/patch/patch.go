package patch

import (
	"fmt"
	"github.com/codecomet-io/installers/sdk/base"
	"github.com/codecomet-io/installers/sdk/base/debian"
	"github.com/codecomet-io/installers/sdk/bin/apt"
	"github.com/codecomet-io/installers/sdk/bin/bash"
	"github.com/codecomet-io/isovaline/sdk/wrapllb"
	"github.com/codecomet-io/isovaline/sdk/wrapllb/platform"
	"github.com/moby/buildkit/client/llb"
	"strings"
)

func Patch(source llb.State, patches ...string) llb.State {
	/*
		grp := &wrapllb.Group{
			ID:           fmt.Sprintf("MACOSSDK-%s-%s", sdkVersion, sdkPath),
			Name:         fmt.Sprintf("Packaging macOS SDK %s from %s", sdkVersion, sdkPath),
			DoNotDisplay: false,
		}

	*/

	// Tooling image with patch
	// XXX this is wrong - this is the host platform, not the worker platform
	toolingImage := base.Debian(debian.Bullseye, platform.DefaultPlatform)
	aptget := apt.New(toolingImage)
	aptget.Group = &wrapllb.Group{
		ID:           "PATCH",
		Name:         "Patching",
		DoNotDisplay: true,
	}
	aptget.Install("patch")
	toolingImage = aptget.State

	mergePatches := []llb.State{}
	command := []string{}
	for k, ptch := range patches {
		mergePatches = append(mergePatches, llb.Scratch().File(llb.Mkfile(fmt.Sprintf("%d.patch", k), 0400, []byte(ptch))))
		command = append(command, fmt.Sprintf("patch -p1 < /patches/%d.patch", k))
	}

	// Start a bash now, readonly, mount the patches and the source
	bsh := bash.New(aptget.State)
	bsh.ReadOnly = true
	bsh.Mount["/patches"] = &wrapllb.State{
		ReadOnly: true,
		Source:   llb.Merge(mergePatches),
	}
	bsh.Mount["/source"] = &wrapllb.State{
		Source: source,
	}
	bsh.Dir = "/source"
	bsh.Run("Patching...", strings.Join(command, "\n"))

	// Return the patched source
	return bsh.Mount["/source"].Source
}
