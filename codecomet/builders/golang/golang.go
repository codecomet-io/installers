package main

import (
	"fmt"
	"github.com/codecomet-io/go-sdk/base/debian"
	"github.com/codecomet-io/go-sdk/base/golang"
	"github.com/codecomet-io/go-sdk/base/llvm"
	"github.com/codecomet-io/go-sdk/base/vcs"
	"github.com/codecomet-io/go-sdk/codecomet"
	"github.com/codecomet-io/go-sdk/codecomet/core"
	"github.com/codecomet-io/go-sdk/controller"
	"github.com/codecomet-io/go-sdk/fileset"
	"github.com/codecomet-io/go-sdk/root"
	"github.com/moby/buildkit/client/llb"
)

var (
	// Variants debian we build
	supportedDebianVersions = []debian.Version{
		debian.Bullseye,
		// debian.Sid,
		// debian.Buster,
		// debian.Bookworm,
		// llvm org packages are not available for bookworm
		// debian.Bookworm,
	}
	// Platforms we build
	/*
		supportedPlatforms = []*codecomet.Platform{
			codecomet.LinuxArm64,
			codecomet.LinuxAmd64,
		}

	*/
	// Versions of LLVM we build
	supportedLLVM = []llvm.Version{
		llvm.V15,
	}

	supportedTargets = []*Target{
		{
			GoVersion:  golang.Go1_20,
			Platform:   codecomet.LinuxArm64,
			GoChecksum: golang.Go1_20DigestArm64,
		},
		{
			GoVersion:  golang.Go1_19,
			Platform:   codecomet.LinuxArm64,
			GoChecksum: golang.Go1_19DigestArm64,
		},
		{
			GoVersion:  golang.Go1_18,
			Platform:   codecomet.LinuxArm64,
			GoChecksum: golang.Go1_18DigestArm64,
		},
		{
			GoVersion:  golang.Go1_20,
			Platform:   codecomet.LinuxAmd64,
			GoChecksum: golang.Go1_20DigestAmd64,
		},
		{
			GoVersion:  golang.Go1_19,
			Platform:   codecomet.LinuxAmd64,
			GoChecksum: golang.Go1_19DigestAmd64,
		},
		{
			GoVersion:  golang.Go1_18,
			Platform:   codecomet.LinuxAmd64,
			GoChecksum: golang.Go1_18DigestAmd64,
		},
	}
)

type Target struct {
	GoVersion  golang.Version
	GoChecksum core.Digest
	Platform   *codecomet.Platform
}

/*
Notes:
- progress group, custom names, are too hard to reach
- mount state is counter-intuitive
- notion of unique identifier for an action is problematic - what truly uniquely define an action, that should be preserved over time?
	- if its the name, we have to make sure you cannot have a duplicate name
	- at least separate display/description/human from unique identifier of the action
- inability to specify FOO=something$FOO is a problem (specifically for PATH) - GetEnv?
- when pushing a resulting image, environment is NOT preserved - likely need to manipulate ImageConfig

Parameterizing plans is WRONG
Components therefore should NOT be plans.
*/

func main() {
	for _, llvmVersion := range supportedLLVM {
		for _, debVersion := range supportedDebianVersions {
			for _, target := range supportedTargets {
				build(target.GoVersion, target.GoChecksum, debVersion, llvmVersion, true, true, target.Platform)
			}
		}
	}
	for _, llvmVersion := range supportedLLVM {
		for _, debVersion := range supportedDebianVersions {
			for _, target := range supportedTargets {
				build(target.GoVersion, target.GoChecksum, debVersion, llvmVersion, true, false, target.Platform)
			}
		}
	}
	for _, llvmVersion := range supportedLLVM {
		for _, debVersion := range supportedDebianVersions {
			for _, target := range supportedTargets {
				build(target.GoVersion, target.GoChecksum, debVersion, llvmVersion, false, false, target.Platform)
			}
		}
	}
}

func build(goVersion golang.Version, goChecksum core.Digest, debianVersion debian.Version, llvmVersion llvm.Version, withCGO bool, withMacOS bool, plt *codecomet.Platform) {
	controller.Init()

	cgo := ""
	mac := ""

	var bb llb.State
	bb = root.Debian(debianVersion, plt).GetInternalState()
	if withCGO {
		cgo = "-cgo"
		bb = root.C(debianVersion, llvmVersion, withMacOS, plt).GetInternalState()
	}
	if withMacOS {
		mac = "-macos"
	}
	bb = vcs.Add(bb)

	outx := fileset.Merge([]codecomet.FileSet{
		fileset.New().Adopt(bb),
		fileset.New().Adopt(golang.Add(goVersion, goChecksum, plt)),
	}, &codecomet.MergeOptions{}).GetInternalState()

	tag := fmt.Sprintf("%s-%s%s%s-%s", debianVersion, goVersion, cgo, mac, plt.Architecture+plt.Variant)

	controller.Get().Exporter = &controller.Export{
		Images: []string{
			"docker.io/codecometio/builder_golang:" + tag,
		},
	}

	controller.Get().Do(outx)
}

// osxbuilda.WithImageConfig()
// maybe on the exporter exptypes.ExporterImageConfigKey: string(bboxConfig), ?
/*
	func buildGo(debianVersion debian.Version, sdkPath string, goVersion golang.Version, osxCrossVersion string, plt *codecomet.Platform) llb.State {
		osxbuilda = codecomet.Execute(osxbuilda, &legacy.Exec{
			Action: codecomet.Action{
				Group: &core.Group{
					DoNotDisplay: true,
				},
			},
			Dir: "/source",
			Args: []string{"bash", "-c", "--", `
					# Should be under the operator control
					echo MACOSX_DEPLOYMENT_TARGET=10.9 >> /etc/profile.d/codecomet_golang
					echo PATH="/macosxcross/bin:$PATH" >> /etc/profile.d/codecomet_golang
					# echo GOENV="$GOENV" >> /etc/profile.d/codecomet_golang
					# go env -w GOCACHE=/golang_cache/GOCACHE GOMODCACHE=/golang_cache/GOMODCACHE
				`},
		})
	}
*/
