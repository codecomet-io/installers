package main

import (
	_ "embed"
	"fmt"
	"github.com/codecomet-io/go-sdk/base/debian"
	golang2 "github.com/codecomet-io/go-sdk/base/golang"
	"github.com/codecomet-io/go-sdk/base/llvm"
	"github.com/codecomet-io/go-sdk/bin/bash"
	"github.com/codecomet-io/go-sdk/bin/golang"
	"github.com/codecomet-io/go-sdk/codecomet"
	"github.com/codecomet-io/go-sdk/codecomet/wrap"
	"github.com/codecomet-io/go-sdk/controller"
	"github.com/codecomet-io/go-sdk/fileset"
	"github.com/codecomet-io/go-sdk/root"
)

/*
Parameterizing plans is WRONG
Components therefore should NOT be plans.
*/

// osxbuilda = osxbuilda.AddEnv("CFLAGS", "-Werror=implicit-function-declaration -Werror=format-security -Wall -O3 -grecord-gcc-switches -g -Wp,-D_GLIBCXX_ASSERTION -D_FORTIFY_SOURCE=2 -pipe -fexceptions -fstack-protector-strong")
// osxbuilda = osxbuilda.AddEnv("CXXFLAGS", "-Werror=implicit-function-declaration -Werror=format-security -Wall -O3 -grecord-gcc-switches -g -Wp,-D_GLIBCXX_ASSERTION -D_FORTIFY_SOURCE=2 -pipe -fexceptions -fstack-protector-strong")

func main() {
	controller.Init()

	// XXX Geez hard coded host path dammmmmm
	outx := fileset.Merge([]codecomet.FileSet{
		buildBuildctl(),
		buildGhost(),
		buildStep(),
		buildIsovaline("/Users/dmp/Projects/GitHub/codecomet/isovaline"),
		buildInternal(),
		buildNerdctl(),
	}, &codecomet.MergeOptions{}).GetInternalState()

	controller.Get().Exporter = &controller.Export{
		Local: "release/mark-I", // os.Args[2],
		// Oci: "oci-tester/exp.tar",
	}

	controller.Get().Do(outx)
}

func buildInternal() codecomet.FileSet {
	version := "v0.10.0"
	src := (&codecomet.Git{
		Reference:  version,
		KeepGitDir: true,
	}).Parse(fmt.Sprintf("https://%s.git", "github.com/github-release/github-release"))

	lamd := buildX(src, "", []*Object{{Name: "linux/amd64/bin/codecomet.internal.github-release", Source: "."}}, codecomet.LinuxAmd64, golang2.Go1_19)
	larm := buildX(src, "", []*Object{{Name: "linux/arm64/bin/codecomet.internal.github-release", Source: "."}}, codecomet.LinuxArm64, golang2.Go1_19)
	armb := buildX(src, "", []*Object{{Name: "codecomet.internal.github-release", Source: "."}}, codecomet.DarwinArm64, golang2.Go1_19)
	amdb := buildX(src, "", []*Object{{Name: "codecomet.internal.github-release", Source: "."}}, codecomet.DarwinAmd64, golang2.Go1_19)
	lip := lipo(armb, amdb, "")

	return fileset.Merge([]codecomet.FileSet{lamd, larm, lip}, &codecomet.MergeOptions{})
}

func buildIsovaline(loc string) codecomet.FileSet {
	version := "mark-I"
	src := &codecomet.Local{
		Path: loc,

		Exclude: []string{
			"xxx",
			"dagger",
			"dist",
			"tmp",
			"tmp_raw",
			".idea",
		},
	}
	lamd := buildX(src, "-X github.com/codecomet-io/isovaline/version.Version="+version, []*Object{
		{Name: "linux/amd64/bin/codecomet-ticker", Source: "./cmd/codecomet-ticker"},
		{Name: "linux/amd64/bin/codecomet-mdns", Source: "./cmd/codecomet-mdns"},
		{Name: "linux/amd64/bin/codecomet-localghost", Source: "./cmd/codecomet-localghost"},
		{Name: "linux/amd64/bin/codecomet-machine", Source: "./cmd/codecomet-machine"},
		{Name: "linux/amd64/bin/codecomet-team", Source: "./cmd/codecomet-team"},
		{Name: "linux/amd64/bin/codecomet-ott", Source: "./cmd/codecomet-ott"},
		{Name: "linux/amd64/bin/codecomet-debugger", Source: "./cmd/codecomet-debugger"},
		{Name: "linux/amd64/bin/codecomet-builder", Source: "./cmd/codecomet-builder"},
		{Name: "linux/amd64/bin/codecomet", Source: "./cmd/codecomet"},
	}, codecomet.LinuxAmd64, golang2.Go1_19)
	larm := buildX(src, "-X github.com/codecomet-io/isovaline/version.Version="+version, []*Object{
		{Name: "linux/arm64/bin/codecomet-ticker", Source: "./cmd/codecomet-ticker"},
		{Name: "linux/arm64/bin/codecomet-mdns", Source: "./cmd/codecomet-mdns"},
		{Name: "linux/arm64/bin/codecomet-localghost", Source: "./cmd/codecomet-localghost"},
		{Name: "linux/arm64/bin/codecomet-machine", Source: "./cmd/codecomet-machine"},
		{Name: "linux/arm64/bin/codecomet-team", Source: "./cmd/codecomet-team"},
		{Name: "linux/arm64/bin/codecomet-ott", Source: "./cmd/codecomet-ott"},
		{Name: "linux/arm64/bin/codecomet-debugger", Source: "./cmd/codecomet-debugger"},
		{Name: "linux/arm64/bin/codecomet-builder", Source: "./cmd/codecomet-builder"},
		{Name: "linux/arm64/bin/codecomet", Source: "./cmd/codecomet"},
	}, codecomet.LinuxArm64, golang2.Go1_19)
	armb := buildX(src, "-X github.com/codecomet-io/isovaline/version.Version="+version, []*Object{
		{Name: "codecomet-ticker", Source: "./cmd/codecomet-ticker"},
		{Name: "codecomet-mdns", Source: "./cmd/codecomet-mdns"},
		{Name: "codecomet-localghost", Source: "./cmd/codecomet-localghost"},
		{Name: "codecomet-machine", Source: "./cmd/codecomet-machine"},
		{Name: "codecomet-team", Source: "./cmd/codecomet-team"},
		{Name: "codecomet-ott", Source: "./cmd/codecomet-ott"},
		{Name: "codecomet-debugger", Source: "./cmd/codecomet-debugger"},
		{Name: "codecomet-builder", Source: "./cmd/codecomet-builder"},
		{Name: "codecomet", Source: "./cmd/codecomet"},
	}, codecomet.DarwinArm64, golang2.Go1_19)
	amdb := buildX(src, "-X github.com/codecomet-io/isovaline/version.Version="+version, []*Object{
		{Name: "codecomet-ticker", Source: "./cmd/codecomet-ticker"},
		{Name: "codecomet-mdns", Source: "./cmd/codecomet-mdns"},
		{Name: "codecomet-localghost", Source: "./cmd/codecomet-localghost"},
		{Name: "codecomet-machine", Source: "./cmd/codecomet-machine"},
		{Name: "codecomet-team", Source: "./cmd/codecomet-team"},
		{Name: "codecomet-ott", Source: "./cmd/codecomet-ott"},
		{Name: "codecomet-debugger", Source: "./cmd/codecomet-debugger"},
		{Name: "codecomet-builder", Source: "./cmd/codecomet-builder"},
		{Name: "codecomet", Source: "./cmd/codecomet"},
	}, codecomet.DarwinAmd64, golang2.Go1_19)
	lip := lipo(armb, amdb, "")

	return fileset.Merge([]codecomet.FileSet{lamd, larm, lip}, &codecomet.MergeOptions{})
}

func buildStep() codecomet.FileSet {
	version := "v0.23.2"
	src := (&codecomet.Git{
		Reference:  version,
		KeepGitDir: true,
	}).Parse(fmt.Sprintf("https://%s.git", "github.com/smallstep/cli"))

	lamd := buildX(src, "-X main.Version=v0.23.2 -X \\\"main.BuildTime=$(date -u '+%Y-%m-%d %H:%M UTC')\\\"", []*Object{{Name: "linux/amd64/bin/codecomet.upstream.step", Source: "./cmd/step"}}, codecomet.LinuxAmd64, golang2.Go1_19)
	larm := buildX(src, "-X main.Version=v0.23.2 -X \\\"main.BuildTime=$(date -u '+%Y-%m-%d %H:%M UTC')\\\"", []*Object{{Name: "linux/arm64/bin/codecomet.upstream.step", Source: "/cmd/step"}}, codecomet.LinuxArm64, golang2.Go1_19)
	armb := buildX(src, "-X main.Version=v0.23.2 -X \\\"main.BuildTime=$(date -u '+%Y-%m-%d %H:%M UTC')\\\"", []*Object{{Name: "codecomet.upstream.step", Source: "/cmd/step"}}, codecomet.DarwinArm64, golang2.Go1_19)
	amdb := buildX(src, "-X main.Version=v0.23.2 -X \\\"main.BuildTime=$(date -u '+%Y-%m-%d %H:%M UTC')\\\"", []*Object{{Name: "codecomet.upstream.step", Source: "/cmd/step"}}, codecomet.DarwinAmd64, golang2.Go1_19)
	lip := lipo(armb, amdb, "")

	return fileset.Merge([]codecomet.FileSet{lamd, larm, lip}, &codecomet.MergeOptions{})
}

func buildGhost() codecomet.FileSet {
	version := "v1.7.1"
	src := (&codecomet.Git{
		Reference:  version,
		KeepGitDir: true,
	}).Parse(fmt.Sprintf("https://%s.git", "github.com/ghostunnel/ghostunnel"))

	lamd := buildX(src, "-X main.version="+version, []*Object{{Name: "linux/amd64/bin/codecomet.upstream.ghostunnel", Source: "."}}, codecomet.LinuxAmd64, golang2.Go1_19)
	larm := buildX(src, "-X main.version="+version, []*Object{{Name: "linux/arm64/bin/codecomet.upstream.ghostunnel", Source: "."}}, codecomet.LinuxArm64, golang2.Go1_19)
	armb := buildX(src, "-X main.version="+version, []*Object{{Name: "codecomet.upstream.ghostunnel", Source: "."}}, codecomet.DarwinArm64, golang2.Go1_19)
	amdb := buildX(src, "-X main.version="+version, []*Object{{Name: "codecomet.upstream.ghostunnel", Source: "."}}, codecomet.DarwinAmd64, golang2.Go1_19)
	lip := lipo(armb, amdb, "")

	return fileset.Merge([]codecomet.FileSet{lamd, larm, lip}, &codecomet.MergeOptions{})
}

func buildBuildctl() codecomet.FileSet {
	version := "v0.11.2"
	src := (&codecomet.Git{
		Reference:  version,
		KeepGitDir: true,
	}).Parse(fmt.Sprintf("https://%s.git", "github.com/moby/buildkit"))

	lamd := buildX(src, "-X github.com/moby/buildkit/version.Version=v0.11.2 -X github.com/moby/buildkit/version.Package=github.com/moby/buildkit", []*Object{{Name: "linux/amd64/bin/codecomet.upstream.buildctl", Source: "./cmd/buildctl"}}, codecomet.LinuxAmd64, golang2.Go1_19)
	larm := buildX(src, "-X github.com/moby/buildkit/version.Version=v0.11.2 -X github.com/moby/buildkit/version.Package=github.com/moby/buildkit", []*Object{{Name: "linux/arm64/bin/codecomet.upstream.buildctl", Source: "./cmd/buildctl"}}, codecomet.LinuxArm64, golang2.Go1_19)
	armb := buildX(src, "-X github.com/moby/buildkit/version.Version=v0.11.2 -X github.com/moby/buildkit/version.Package=github.com/moby/buildkit", []*Object{{Name: "codecomet.upstream.buildctl", Source: "./cmd/buildctl"}}, codecomet.DarwinArm64, golang2.Go1_19)
	amdb := buildX(src, "-X github.com/moby/buildkit/version.Version=v0.11.2 -X github.com/moby/buildkit/version.Package=github.com/moby/buildkit", []*Object{{Name: "codecomet.upstream.buildctl", Source: "./cmd/buildctl"}}, codecomet.DarwinAmd64, golang2.Go1_19)
	lip := lipo(armb, amdb, "")

	return fileset.Merge([]codecomet.FileSet{lamd, larm, lip}, &codecomet.MergeOptions{})
}

func buildNerdctl() codecomet.FileSet {
	version := "v1.2.0"
	src := (&codecomet.Git{
		Reference:  version,
		KeepGitDir: true,
	}).Parse(fmt.Sprintf("https://%s.git", "github.com/containerd/nerdctl"))

	lamd := buildX(src, "-X github.com/containerd/nerdctl/pkg/version.Version="+version, []*Object{{Name: "linux/amd64/bin/codecomet.upstream.nerdctl", Source: "./cmd/nerdctl"}}, codecomet.LinuxAmd64, golang2.Go1_19)
	larm := buildX(src, "-X github.com/containerd/nerdctl/pkg/version.Version="+version, []*Object{{Name: "linux/arm64/bin/codecomet.upstream.nerdctl", Source: "./cmd/nerdctl"}}, codecomet.LinuxArm64, golang2.Go1_19)
	// armb := buildX(source, "-X github.com/containerd/nerdctl/pkg/version.Version="+version, []*Object{{Name: "codecomet.upstream.nerdctl", Source: "./cmd/nerdctl"}}, codecomet.DarwinArm64, golang2.Go1_19)
	// amdb := buildX(source, "-X github.com/containerd/nerdctl/pkg/version.Version="+version, []*Object{{Name: "codecomet.upstream.nerdctl", Source: "./cmd/nerdctl"}}, codecomet.DarwinAmd64, golang2.Go1_19)
	// lip := lipo(armb, amdb, "")

	return fileset.Merge([]codecomet.FileSet{lamd, larm}, &codecomet.MergeOptions{})
}

type Object struct {
	Repo    string
	Package string
	Version string
	Source  string
	Name    string
}

func lipo(armbuild codecomet.FileSet, amdbuild codecomet.FileSet, entitlements string) codecomet.FileSet {
	// Need lipo - could be a different image?
	bsh := bash.New(root.C(debian.Bullseye, llvm.V15, true, codecomet.DefaultPlatform).GetInternalState())
	bsh.ReadOnly = true
	bsh.Mount["/armbuild"] = &wrap.State{
		Source:   armbuild.GetInternalState(),
		ReadOnly: true,
	}
	bsh.Mount["/amdbuild"] = &wrap.State{
		Source:   amdbuild.GetInternalState(),
		ReadOnly: true,
	}
	bsh.Mount["/output"] = &wrap.State{
		Source: fileset.New().GetInternalState(),
	}
	bsh.Env["entitlements"] = entitlements

	bsh.Run("LIPO", `
		args=("-f")
		[ "$entitlements" != "" ] || {
			echo "$entitlements" > entitlements.plist
			args+=(--entitlements entitlements.plist)
		}

		mkdir -p /output/darwin/universal/bin
		for i in /armbuild/*; do
			/opt/macosxcross/bin/lipo -create -output /output/darwin/universal/bin/$(basename "$i") /armbuild/$(basename "$i") /amdbuild/$(basename "$i")
			# Resign?
			PATH=/opt/macosxcross/bin:$PATH LD_LIBRARY_PATH=/opt/macosxcross/bin codesign "${args[@]}" --sign - /output/darwin/universal/bin/$(basename "$i")
		done
`)
	scr := fileset.New()
	scr.Adopt(bsh.Mount["/output"].Source)
	return scr
}

func buildX(src codecomet.FileSet, verpoint string, targets []*Object, plt *codecomet.Platform, golangVersion golang2.Version, patches ...string) codecomet.FileSet {
	if len(patches) > 0 {
		src.Patch(patches, &codecomet.PatchOptions{})
		// src.Adopt(patch.Patch(src.GetInternalState(), patches...))
	}

	glg := golang.New(debian.Bullseye, golangVersion, true, true)
	glg.Source = src.GetInternalState()

	glg.Env["RELEASE"] = "true"

	glg.Env["MACOSX_DEPLOYMENT_TARGET"] = "11.0"

	glg.Config.Target.GOOS = golang.GoOS(plt.OS)
	glg.Config.Target.GOARCH = golang.GoArch(plt.Architecture)
	glg.Config.CGO.CGO_ENABLED = 1
	glg.Config.CGO.GO_EXTLINK_ENABLED = 1

	command := `
		gcflags=
		ldflags="-w -s"
		[ "$RELEASE" ] || {
			gcflags="all=-N -l"
			ldflags=
		}
	`

	// XXX just because of github-releaser
	for _, v := range targets {
		command += fmt.Sprintf(`
			output=%q
			mkdir -p $(dirname /output/$output)
			recheat=/input
			[ ! -e "/input/Gopkg.lock" ] || {
				cp -R /input /codecomet/source
				cd /codecomet/source
				# GEEEEZ
				go mod init github.com/github-release/github-release
				go mod tidy
				go mod vendor
				recheat=/codecomet/source
			}
			go build -o /output/$output -tags "cgo netcgo osusergo" -gcflags "$gcflags" -ldflags "$ldflags %s" $recheat/%q
`, v.Name, verpoint, v.Source)
	}

	res := glg.Do(command)
	scr := fileset.New()
	scr.Adopt(res)
	return scr
}
