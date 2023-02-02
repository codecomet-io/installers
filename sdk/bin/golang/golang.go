package golang

import (
	"fmt"
	"github.com/codecomet-io/installers/sdk/base"
	"github.com/codecomet-io/installers/sdk/base/debian"
	"github.com/codecomet-io/installers/sdk/base/golang"
	"github.com/codecomet-io/installers/sdk/bin/bash"
	"github.com/codecomet-io/isovaline/sdk/wrapllb"
	"github.com/codecomet-io/isovaline/sdk/wrapllb/platform"
	"github.com/moby/buildkit/client/llb"
	"strings"
)

const (
	sharedStoreMountLocation = "/_cc/share/go"
)

var (
	sharedStore = &wrapllb.Cache{
		UniqueDescription: fmt.Sprintf("golang shared cache 2"),
		SharingMode:       wrapllb.ModeShared,
	}
)

type Go struct {
	// Bash embed
	bash.Bash

	// Inherit config object
	Expert *Config

	Source      llb.State
	Destination llb.State

	// Base env
	// Env map[string]string

	// State we are manipulating
	// State llb.State

}

func New(debianVersion debian.Version, golangVersion golang.Version, withCGO bool, withMacOS bool) *Go {
	st := base.Go(debianVersion, golangVersion, withCGO, withMacOS, platform.DefaultPlatform)
	// XXX wrong platform
	ap := &Go{
		Bash:        *bash.New(st),
		Expert:      &Config{},
		Destination: llb.Scratch(),
	}

	ap.Bash.Cache[sharedStoreMountLocation] = sharedStore
	ap.Bash.Dir = "/input"
	ap.Bash.ReadOnly = true

	return ap
}

func (g *Go) Do(args ...string) llb.State {
	com := strings.Join(args, "\n")
	name := com
	/*
		com := "go "

		for _, v := range args {
			com = fmt.Sprintf("%s %q", com, v)
		}

	*/

	for k, v := range g.Expert.ToEnv() {
		g.Bash.Env[k] = v
	}

	g.Bash.Env["GOENV"] = "off"
	g.Bash.Env["GOPATH"] = sharedStoreMountLocation + "/GOPATH"
	g.Bash.Env["GOCACHE"] = sharedStoreMountLocation + "/GOCACHE"
	g.Bash.Env["GOMODCACHE"] = sharedStoreMountLocation + "/GOMODCACHE"

	g.Bash.Env["PATH"] = "/opt/macosxcross/bin:/opt/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"
	// XXX move to config
	g.Bash.Env["MACOSX_DEPLOYMENT_TARGET"] = "10.9"

	if g.Expert.CGO_ENABLED == Enabled {
		g.Bash.Env["CFLAGS"] = ""   // XXX CFLAGS
		g.Bash.Env["CXXFLAGS"] = "" // XXX CXXFLAGS

		if g.Expert.GOOS == Darwin {
			/*
			   #0 212.6 All done! Now you can use o32-clang(++) and o64-clang(++) like a normal compiler.
			   #0 212.6 !!! Use aarch64-apple-darwin21.2-* instead of arm64-* when dealing with Automake !!!
			   #0 212.6 !!! CC=aarch64-apple-darwin21.2-clang ./configure --host=aarch64-apple-darwin21.2 !!!
			   #0 212.6 !!! CC="aarch64-apple-darwin21.2-clang -arch arm64e" ./configure --host=aarch64-apple-darwin21.2 !!!
			*/
			// clang llvm
			if g.Expert.GOARCH == Amd64 {
				g.Bash.Env["CC"] = "o64-clang"
				g.Bash.Env["CXX"] = "o64-clang++"
			} else if g.Expert.GOARCH == Arm64 {
				// envs["CC"] = "aarch64-apple-darwin22-clang"
				// envs["CXX"] = "aarch64-apple-darwin22-clang++"
				// XXX this will break with SDK updates - this digit is dependent on the SDK version
				g.Bash.Env["CC"] = "arm64e-apple-darwin22.2-clang" // arm64e-apple-darwin22.2-clang++
				g.Bash.Env["CXX"] = "arm64e-apple-darwin22.2-clang++"
			}
			g.Bash.Env["CFLAGS"] = ""   // XXX envs["CFLAGS"] + " " + COMPILER_OPTIONS_DARWIN
			g.Bash.Env["CXXFLAGS"] = "" // XXX envs["CXXFLAGS"] + " " + COMPILER_OPTIONS_DARWIN
		} else {
			g.Bash.Env["CFLAGS"] = g.Bash.Env["CFLAGS"] + " " + COMPILER_OPTIONS_LINUX
			g.Bash.Env["CXXFLAGS"] = g.Bash.Env["CXXFLAGS"] + " " + COMPILER_OPTIONS_LINUX
			// https://gcc.gnu.org/onlinedocs/gcc/Link-Options.html#Link-Options
			// Note: none of these are working for macOS linker
			DEB_TARGET_GNU_TYPE := ""
			if g.Expert.GOARCH == Amd64 {
				DEB_TARGET_GNU_TYPE = "x86_64-linux-gnu"
				g.Bash.Env["CFLAGS"] = g.Bash.Env["CFLAGS"] + " " + COMPILER_OPTIONS_LINUX_AMD64
			} else if g.Expert.GOARCH == Arm64 {
				DEB_TARGET_GNU_TYPE = "aarch64-linux-gnu"
			} else if g.Expert.GOARCH == Arm {
				DEB_TARGET_GNU_TYPE = "arm-linux-gnueabihf"
			}
			g.Bash.Env["PKG_CONFIG"] = DEB_TARGET_GNU_TYPE + "-pkg-config"
			g.Bash.Env["AR"] = DEB_TARGET_GNU_TYPE + "-ar"
			g.Bash.Env["CC"] = DEB_TARGET_GNU_TYPE + "-gcc"
			g.Bash.Env["CXX"] = DEB_TARGET_GNU_TYPE + "-g++"

			g.Bash.Env["LDFLAGS"] = "-Wl,-z,relro -Wl,-z,now -Wl,-z,defs -Wl,-z,noexecstack"
		}
	}

	// var pg *wrapllb.Group
	g.Bash.Mount["/input"] = &wrapllb.State{
		// XXX lima does insist on writing under _output
		ReadOnly: true,
		NoOutput: true,
		Source:   g.Source,
		Path:     "/",
	}
	g.Bash.Mount["/output"] = &wrapllb.State{
		Source: g.Destination,
	}

	g.Bash.Run(name, com)

	g.Destination = g.Bash.Mount["/output"].Source
	return g.Destination
}
