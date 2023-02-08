package main

import (
	_ "embed"
	"fmt"
	"github.com/codecomet-io/go-sdk/base"
	"github.com/codecomet-io/go-sdk/base/debian"
	golang2 "github.com/codecomet-io/go-sdk/base/golang"
	"github.com/codecomet-io/go-sdk/base/llvm"
	"github.com/codecomet-io/go-sdk/bin/bash"
	"github.com/codecomet-io/go-sdk/bin/golang"
	"github.com/codecomet-io/go-sdk/bin/patch"
	"github.com/codecomet-io/go-sdk/codecomet"
	"github.com/codecomet-io/go-sdk/controller"
	"github.com/codecomet-io/go-sdk/coretypes"
	"github.com/codecomet-io/go-sdk/wrapllb"
	"github.com/moby/buildkit/client/llb"
)

/*
Parameterizing plans is WRONG
Components therefore should NOT be plans.
*/

// osxbuilda = osxbuilda.AddEnv("CFLAGS", "-Werror=implicit-function-declaration -Werror=format-security -Wall -O3 -grecord-gcc-switches -g -Wp,-D_GLIBCXX_ASSERTION -D_FORTIFY_SOURCE=2 -pipe -fexceptions -fstack-protector-strong")
// osxbuilda = osxbuilda.AddEnv("CXXFLAGS", "-Werror=implicit-function-declaration -Werror=format-security -Wall -O3 -grecord-gcc-switches -g -Wp,-D_GLIBCXX_ASSERTION -D_FORTIFY_SOURCE=2 -pipe -fexceptions -fstack-protector-strong")

func main() {
	codecomet.Init()

	// XXX this is largely incorrect - these are GoLang platforms, not OCI
	// XXX Geez hard coded host path dammmmmm
	outx := build("/Users/dmp/Projects/GitHub/codecomet/isovaline", []*coretypes.Platform{
		coretypes.LinuxArm64,
		coretypes.LinuxAmd64,
		coretypes.DarwinArm64,
		coretypes.DarwinAmd64,
	})

	// output = wrapllb.Copy(upstream, "/dist/*", output, "/", &wrapllb.CopyOptions{})
	controller.Get().Exporter = &controller.Export{
		Local: "release/mark-I", // os.Args[2],
		// Oci: "oci-tester/exp.tar",
	}

	controller.Get().Do(outx)
}

func buildLima(loc string, plt []*coretypes.Platform) llb.State {
	source := codecomet.From(&codecomet.Local{
		Path: loc,

		Exclude: []string{
			"xxx",
			"dagger",
			"dist",
			"tmp",
			"tmp_raw",
			".idea",
		},
	})

	// source = patch.Patch(source, lima_patch, lima_sdk)

	outx := []llb.State{}

	for _, v := range plt {
		outy := codecomet.From(&codecomet.Scratch{})
		glg := golang.New(debian.Bullseye, golang2.Go1_19, true, true)
		glg.Source = source
		// Isovaline destination
		glg.Env["DC_PREFIX"] = "/output"
		// XXX not quite sure what is going on with the SDK at this point
		// glg.Env["MACOSX_DEPLOYMENT_TARGET"] = "11.0"

		// glg.Env["SYSTEM_VERSION_COMPAT"] = "1"

		glg.Config.Target.GOOS = golang.GoOS(v.OS)
		glg.Config.Target.GOARCH = golang.GoArch(v.Architecture)

		glg.Do(`
			LIMA_ROOT=/input/dependencies/lima-vm/lima
			mkdir -p "$DC_PREFIX"/bin
			mkdir -p "$DC_PREFIX"/share/lima/examples

			#cp -a "$LIMA_ROOT"/cmd/apptainer.lima "$DC_PREFIX"/bin
			#cp -a "$LIMA_ROOT"/cmd/docker.lima "$DC_PREFIX"/bin
			cp -a "$LIMA_ROOT"/cmd/lima "$DC_PREFIX"/bin
			#cp -a "$LIMA_ROOT"/cmd/lima.bat "$DC_PREFIX"/bin
			cp -a "$LIMA_ROOT"/cmd/nerdctl.lima "$DC_PREFIX"/bin
			#cp -a "$LIMA_ROOT"/cmd/podman.lima "$DC_PREFIX"/bin

			# Link our template
			cp /input/upstream/lima-cli/templates/codecomet-bullseye.yaml "$DC_PREFIX"/share/lima/examples
			ln -s codecomet-buildkit-debian.yaml "$DC_PREFIX"/share/lima/examples/default.yaml
		`)

		glg.Do(`
			LIMA_ROOT=/input/dependencies/lima-vm/lima
			cd "$LIMA_ROOT"
			GOOS=linux GOARCH=arm64 go build -o "$DC_PREFIX"/share/lima/lima-guestagent.Linux-aarch64 ./cmd/lima-guestagent
			GOOS=linux GOARCH=amd64 go build -o "$DC_PREFIX"/share/lima/lima-guestagent.Linux-x86_64 ./cmd/lima-guestagent
		`)

		glg.Config.CGO.CC = "clang"
		glg.Config.CGO.CXX = "clang++"
		glg.Config.CGO.CGO_ENABLED = 1
		glg.Config.CGO.GO_EXTLINK_ENABLED = 1

		glg.Do(
			`
				LIMA_ROOT=/input/dependencies/lima-vm/lima
				cd "$LIMA_ROOT"
				go build -o "$DC_PREFIX"/bin/limactl ./cmd/limactl
			`,
		)

		if glg.Config.Target.GOOS == golang.Darwin {
			glg.Do(`
				# XXX Very dirty right now - fix this!
				# LD_LIBRARY_PATH=/opt/macosxcross/bin codesign -f --sign - 
				LIMA_ROOT=/input/dependencies/lima-vm/lima
				codesign --entitlements "$LIMA_ROOT"/vz.entitlements --sign - "$DC_PREFIX"/bin/limactl
			`)
		}

		// fmt.Println(v, cc)
		outx = append(outx, wrapllb.Copy(glg.Destination, "/", outy, "/"+v.OS+"/"+v.Architecture, &wrapllb.CopyOptions{}))
	}

	/*
		CGO_ENABLED=1 $(GO_BUILD) -o bin/limactl ./cmd/limactl

		GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GO_BUILD) -o _output/share/lima/lima-guestagent.Linux-x86_64 ./cmd/lima-guestagent
		GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GO_BUILD) -o _output/share/lima/lima-guestagent.Linux-x86_64 ./cmd/lima-guestagent

				cp -aL examples _output/share/lima
				mkdir -p _output/share/doc/lima
				cp -aL *.md LICENSE docs _output/share/doc/lima
	*/
	return llb.Merge(outx)
}

func build(loc string, plt []*coretypes.Platform) llb.State {
	out := []llb.State{}
	out = append(out, buildBuildctl())
	out = append(out, buildGhost())
	out = append(out, buildStep())
	out = append(out, buildIsovaline(loc))
	out = append(out, buildInternal())
	out = append(out, buildNerdctl())
	return llb.Merge(out)
}

/*
func buildLima() llb.State {
	version := "v0.14.2"
	source := codecomet.From((&codecomet.Git{
		Reference:  version,
		KeepGitDir: true,
	}).Parse(fmt.Sprintf("https://%s.git", "github.com/lima-vm/lima")))

	lamd := buildX(source, "", []*Object{{Name: "linux/amd64/bin/codecomet.upstream.limactl", Source: "."}}, coretypes.LinuxAmd64, golang2.Go1_19)
	larm := buildX(source, "", []*Object{{Name: "linux/arm64/bin/codecomet.internal.github-release", Source: "."}}, coretypes.LinuxArm64, golang2.Go1_19)
	armb := buildX(source, "", []*Object{{Name: "codecomet.internal.github-release", Source: "."}}, coretypes.DarwinArm64, golang2.Go1_19)
	amdb := buildX(source, "", []*Object{{Name: "codecomet.internal.github-release", Source: "."}}, coretypes.DarwinAmd64, golang2.Go1_19)
	lip := lipo(armb, amdb, "")

	return llb.Merge([]llb.State{
		lamd,
		larm,
		lip,
	})
}
*/

func buildInternal() llb.State {
	version := "v0.10.0"
	source := codecomet.From((&codecomet.Git{
		Reference:  version,
		KeepGitDir: true,
	}).Parse(fmt.Sprintf("https://%s.git", "github.com/github-release/github-release")))

	lamd := buildX(source, "", []*Object{{Name: "linux/amd64/bin/codecomet.internal.github-release", Source: "."}}, coretypes.LinuxAmd64, golang2.Go1_19)
	larm := buildX(source, "", []*Object{{Name: "linux/arm64/bin/codecomet.internal.github-release", Source: "."}}, coretypes.LinuxArm64, golang2.Go1_19)
	armb := buildX(source, "", []*Object{{Name: "codecomet.internal.github-release", Source: "."}}, coretypes.DarwinArm64, golang2.Go1_19)
	amdb := buildX(source, "", []*Object{{Name: "codecomet.internal.github-release", Source: "."}}, coretypes.DarwinAmd64, golang2.Go1_19)
	lip := lipo(armb, amdb, "")

	return llb.Merge([]llb.State{
		lamd,
		larm,
		lip,
	})
}

func buildIsovaline(loc string) llb.State {
	version := "mark-I"
	source := codecomet.From(&codecomet.Local{
		Path: loc,

		Exclude: []string{
			"xxx",
			"dagger",
			"dist",
			"tmp",
			"tmp_raw",
			".idea",
		},
	})

	lamd := buildX(source, "-X github.com/codecomet-io/isovaline/version.Version="+version, []*Object{
		{Name: "linux/amd64/bin/codecomet-ticker", Source: "./cmd/codecomet-ticker"},
		{Name: "linux/amd64/bin/codecomet-mdns", Source: "./cmd/codecomet-mdns"},
		{Name: "linux/amd64/bin/codecomet-localghost", Source: "./cmd/codecomet-localghost"},
		{Name: "linux/amd64/bin/codecomet-machine", Source: "./cmd/codecomet-machine"},
		{Name: "linux/amd64/bin/codecomet-team", Source: "./cmd/codecomet-team"},
		{Name: "linux/amd64/bin/codecomet-ott", Source: "./cmd/codecomet-ott"},
		{Name: "linux/amd64/bin/codecomet-debugger", Source: "./cmd/codecomet-debugger"},
		{Name: "linux/amd64/bin/codecomet-builder", Source: "./cmd/codecomet-builder"},
	}, coretypes.LinuxAmd64, golang2.Go1_19)
	larm := buildX(source, "-X github.com/codecomet-io/isovaline/version.Version="+version, []*Object{
		{Name: "linux/arm64/bin/codecomet-ticker", Source: "./cmd/codecomet-ticker"},
		{Name: "linux/arm64/bin/codecomet-mdns", Source: "./cmd/codecomet-mdns"},
		{Name: "linux/arm64/bin/codecomet-localghost", Source: "./cmd/codecomet-localghost"},
		{Name: "linux/arm64/bin/codecomet-machine", Source: "./cmd/codecomet-machine"},
		{Name: "linux/arm64/bin/codecomet-team", Source: "./cmd/codecomet-team"},
		{Name: "linux/arm64/bin/codecomet-ott", Source: "./cmd/codecomet-ott"},
		{Name: "linux/arm64/bin/codecomet-debugger", Source: "./cmd/codecomet-debugger"},
		{Name: "linux/arm64/bin/codecomet-builder", Source: "./cmd/codecomet-builder"},
	}, coretypes.LinuxArm64, golang2.Go1_19)
	armb := buildX(source, "-X github.com/codecomet-io/isovaline/version.Version="+version, []*Object{
		{Name: "codecomet-ticker", Source: "./cmd/codecomet-ticker"},
		{Name: "codecomet-mdns", Source: "./cmd/codecomet-mdns"},
		{Name: "codecomet-localghost", Source: "./cmd/codecomet-localghost"},
		{Name: "codecomet-machine", Source: "./cmd/codecomet-machine"},
		{Name: "codecomet-team", Source: "./cmd/codecomet-team"},
		{Name: "codecomet-ott", Source: "./cmd/codecomet-ott"},
		{Name: "codecomet-debugger", Source: "./cmd/codecomet-debugger"},
		{Name: "codecomet-builder", Source: "./cmd/codecomet-builder"},
	}, coretypes.DarwinArm64, golang2.Go1_19)
	amdb := buildX(source, "-X github.com/codecomet-io/isovaline/version.Version="+version, []*Object{
		{Name: "codecomet-ticker", Source: "./cmd/codecomet-ticker"},
		{Name: "codecomet-mdns", Source: "./cmd/codecomet-mdns"},
		{Name: "codecomet-localghost", Source: "./cmd/codecomet-localghost"},
		{Name: "codecomet-machine", Source: "./cmd/codecomet-machine"},
		{Name: "codecomet-team", Source: "./cmd/codecomet-team"},
		{Name: "codecomet-ott", Source: "./cmd/codecomet-ott"},
		{Name: "codecomet-debugger", Source: "./cmd/codecomet-debugger"},
		{Name: "codecomet-builder", Source: "./cmd/codecomet-builder"},
	}, coretypes.DarwinAmd64, golang2.Go1_19)
	lip := lipo(armb, amdb, "")

	return llb.Merge([]llb.State{
		lamd,
		larm,
		lip,
	})
}

func buildStep() llb.State {
	version := "v0.23.2"
	source := codecomet.From((&codecomet.Git{
		Reference:  version,
		KeepGitDir: true,
	}).Parse(fmt.Sprintf("https://%s.git", "github.com/smallstep/cli")))

	lamd := buildX(source, "-X main.Version=v0.23.2 -X \\\"main.BuildTime=$(date -u '+%Y-%m-%d %H:%M UTC')\\\"", []*Object{{Name: "linux/amd64/bin/codecomet.upstream.step", Source: "./cmd/step"}}, coretypes.LinuxAmd64, golang2.Go1_19)
	larm := buildX(source, "-X main.Version=v0.23.2 -X \\\"main.BuildTime=$(date -u '+%Y-%m-%d %H:%M UTC')\\\"", []*Object{{Name: "linux/arm64/bin/codecomet.upstream.step", Source: "/cmd/step"}}, coretypes.LinuxArm64, golang2.Go1_19)
	armb := buildX(source, "-X main.Version=v0.23.2 -X \\\"main.BuildTime=$(date -u '+%Y-%m-%d %H:%M UTC')\\\"", []*Object{{Name: "codecomet.upstream.step", Source: "/cmd/step"}}, coretypes.DarwinArm64, golang2.Go1_19)
	amdb := buildX(source, "-X main.Version=v0.23.2 -X \\\"main.BuildTime=$(date -u '+%Y-%m-%d %H:%M UTC')\\\"", []*Object{{Name: "codecomet.upstream.step", Source: "/cmd/step"}}, coretypes.DarwinAmd64, golang2.Go1_19)
	lip := lipo(armb, amdb, "")

	return llb.Merge([]llb.State{
		lamd,
		larm,
		lip,
	})
}

func buildGhost() llb.State {
	version := "v1.7.1"
	source := codecomet.From((&codecomet.Git{
		Reference:  version,
		KeepGitDir: true,
	}).Parse(fmt.Sprintf("https://%s.git", "github.com/ghostunnel/ghostunnel")))

	lamd := buildX(source, "-X main.version="+version, []*Object{{Name: "linux/amd64/bin/codecomet.upstream.ghostunnel", Source: "."}}, coretypes.LinuxAmd64, golang2.Go1_19)
	larm := buildX(source, "-X main.version="+version, []*Object{{Name: "linux/arm64/bin/codecomet.upstream.ghostunnel", Source: "."}}, coretypes.LinuxArm64, golang2.Go1_19)
	armb := buildX(source, "-X main.version="+version, []*Object{{Name: "codecomet.upstream.ghostunnel", Source: "."}}, coretypes.DarwinArm64, golang2.Go1_19)
	amdb := buildX(source, "-X main.version="+version, []*Object{{Name: "codecomet.upstream.ghostunnel", Source: "."}}, coretypes.DarwinAmd64, golang2.Go1_19)
	lip := lipo(armb, amdb, "")

	return llb.Merge([]llb.State{
		lamd,
		larm,
		lip,
	})
}

func buildBuildctl() llb.State {
	version := "v0.11.2"
	source := codecomet.From((&codecomet.Git{
		Reference:  version,
		KeepGitDir: true,
	}).Parse(fmt.Sprintf("https://%s.git", "github.com/moby/buildkit")))

	lamd := buildX(source, "-X github.com/moby/buildkit/version.Version=v0.11.2 -X github.com/moby/buildkit/version.Package=github.com/moby/buildkit", []*Object{{Name: "linux/amd64/bin/codecomet.upstream.buildctl", Source: "./cmd/buildctl"}}, coretypes.LinuxAmd64, golang2.Go1_19)
	larm := buildX(source, "-X github.com/moby/buildkit/version.Version=v0.11.2 -X github.com/moby/buildkit/version.Package=github.com/moby/buildkit", []*Object{{Name: "linux/arm64/bin/codecomet.upstream.buildctl", Source: "./cmd/buildctl"}}, coretypes.LinuxArm64, golang2.Go1_19)
	armb := buildX(source, "-X github.com/moby/buildkit/version.Version=v0.11.2 -X github.com/moby/buildkit/version.Package=github.com/moby/buildkit", []*Object{{Name: "codecomet.upstream.buildctl", Source: "./cmd/buildctl"}}, coretypes.DarwinArm64, golang2.Go1_19)
	amdb := buildX(source, "-X github.com/moby/buildkit/version.Version=v0.11.2 -X github.com/moby/buildkit/version.Package=github.com/moby/buildkit", []*Object{{Name: "codecomet.upstream.buildctl", Source: "./cmd/buildctl"}}, coretypes.DarwinAmd64, golang2.Go1_19)
	lip := lipo(armb, amdb, "")

	return llb.Merge([]llb.State{
		lamd,
		larm,
		lip,
	})
}

func buildNerdctl() llb.State {
	version := "v1.2.0"
	source := codecomet.From((&codecomet.Git{
		Reference:  version,
		KeepGitDir: true,
	}).Parse(fmt.Sprintf("https://%s.git", "github.com/containerd/nerdctl")))

	lamd := buildX(source, "-X github.com/containerd/nerdctl/pkg/version.Version="+version, []*Object{{Name: "linux/amd64/bin/codecomet.upstream.nerdctl", Source: "./cmd/nerdctl"}}, coretypes.LinuxAmd64, golang2.Go1_19)
	larm := buildX(source, "-X github.com/containerd/nerdctl/pkg/version.Version="+version, []*Object{{Name: "linux/arm64/bin/codecomet.upstream.nerdctl", Source: "./cmd/nerdctl"}}, coretypes.LinuxArm64, golang2.Go1_19)
	// armb := buildX(source, "-X github.com/containerd/nerdctl/pkg/version.Version="+version, []*Object{{Name: "codecomet.upstream.nerdctl", Source: "./cmd/nerdctl"}}, coretypes.DarwinArm64, golang2.Go1_19)
	// amdb := buildX(source, "-X github.com/containerd/nerdctl/pkg/version.Version="+version, []*Object{{Name: "codecomet.upstream.nerdctl", Source: "./cmd/nerdctl"}}, coretypes.DarwinAmd64, golang2.Go1_19)
	// lip := lipo(armb, amdb, "")

	return llb.Merge([]llb.State{
		lamd,
		larm,
		// lip,
	})
}

type Object struct {
	Repo    string
	Package string
	Version string
	Source  string
	Name    string
}

func lipo(armbuild llb.State, amdbuild llb.State, entitlements string) llb.State {
	// Need lipo - could be a different image?
	bsh := bash.New(base.C(debian.Bullseye, llvm.V15, true, coretypes.DefaultPlatform))
	bsh.ReadOnly = true
	bsh.Mount["/armbuild"] = &wrapllb.State{
		Source:   armbuild,
		ReadOnly: true,
	}
	bsh.Mount["/amdbuild"] = &wrapllb.State{
		Source:   amdbuild,
		ReadOnly: true,
	}
	bsh.Mount["/output"] = &wrapllb.State{
		Source: codecomet.From(&codecomet.Scratch{}),
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
	return bsh.Mount["/output"].Source
}

func buildX(source llb.State, verpoint string, targets []*Object, plt *coretypes.Platform, golangVersion golang2.Version, patches ...string) llb.State {
	if len(patches) > 0 {
		source = patch.Patch(source, patches...)
	}

	glg := golang.New(debian.Bullseye, golangVersion, true, true)
	glg.Source = source

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
	return glg.Do(command)
}
