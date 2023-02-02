package vcs

import (
	"github.com/codecomet-io/installers/sdk/bin/apt"
	"github.com/moby/buildkit/client/llb"
)

func Add(debianState llb.State) llb.State {
	aptGet := apt.New(debianState)
	aptGet.Install("git", "mercurial")
	return aptGet.State
}

func Tier2(debianState llb.State) llb.State {
	aptGet := apt.New(debianState)
	aptGet.Install("bzr", "subversion", "cvs", "rcs")
	return aptGet.State
}
