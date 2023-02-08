package main

import (
	_ "embed"
	"fmt"
	"github.com/codecomet-io/go-sdk/base/debian"
	golang2 "github.com/codecomet-io/go-sdk/base/golang"
	"github.com/codecomet-io/go-sdk/bin/golang"
	"github.com/codecomet-io/go-sdk/codecomet"
	"github.com/codecomet-io/go-sdk/controller"
	"github.com/codecomet-io/go-sdk/coretypes"
	"github.com/codecomet-io/go-sdk/wrapllb"
	"github.com/moby/buildkit/client/llb"
)

//go:embed templates/codecomet-bullseye.yaml
var vmtpl string

//go:embed entitlements/lima.plist
var entitl string

func main() {
	codecomet.Init()

	// XXX this is largely incorrect - these are GoLang platforms, not OCI
	// XXX Geez hard coded host path dammmmmm
	outx := build()

	// output = wrapllb.Copy(upstream, "/dist/*", output, "/", &wrapllb.CopyOptions{})
	controller.Get().Exporter = &controller.Export{
		Local: "release/mark-I", // os.Args[2],
		// Oci: "oci-tester/exp.tar",
	}

	controller.Get().Do(outx)
}

func buildone(source llb.State, repo string, version string, entitlements string, v *coretypes.Platform) llb.State {

	ent := codecomet.From(&codecomet.Scratch{}).File(llb.Mkfile("entitlements.plist", 0400, []byte(entitlements)))

	outy := codecomet.From(&codecomet.Scratch{})
	glg := golang.New(debian.Bullseye, golang2.Go1_19, true, true)
	glg.Source = source

	glg.Destination = glg.Destination.
		File(llb.Mkdir("/bin", 0700)).
		File(llb.Mkdir("/share/lima/examples", 0700, llb.WithParents(true))).
		File(llb.Mkfile("/share/lima/examples/codecomet-bullseye.yaml", 0600, []byte(vmtpl)))
	// These are mere shell wrapping a call to limactl
	// File(llb.Copy(glg.Source, "/cmd/lima", "/bin", &llb.CopyInfo{})).
	// File(llb.Copy(glg.Source, "/cmd/nerdctl.lima", "/bin", &llb.CopyInfo{})).
	// apptainer
	// docker
	// lima
	// podman
	// symlinking TBD
	// File(llb.Mkfile("/share/lima/examples/default.yaml", fs.ModeSymlink|0600, []byte("codecomet-bullseye.yaml"), &llb.MkfileInfo{}))

	glg.Env["DESTINATION"] = "/output"
	glg.Env["LIMA_ROOT"] = "/input"
	glg.Env["PACKAGE"] = repo
	glg.Env["VERSION"] = version

	// XXX should this be in go instead?
	glg.Env["MACOSX_DEPLOYMENT_TARGET"] = "13.0"

	glg.Do(`
		ln -s codecomet-bullseye.yaml "$DESTINATION"/share/lima/examples/default.yaml
	`)

	glg.Config.Target.GOOS = golang.Linux
	glg.Config.Target.GOARCH = golang.Amd64
	glg.Do(`
		cd "$LIMA_ROOT"
		go build -ldflags="-s -w -X "$PACKAGE"/pkg/version.Version=$VERSION" -o "$DESTINATION"/share/lima/lima-guestagent.Linux-x86_64 ./cmd/lima-guestagent
	`)

	glg.Config.Target.GOARCH = golang.Arm64
	glg.Do(`
		cd "$LIMA_ROOT"
		go build -ldflags="-s -w -X "$PACKAGE"/pkg/version.Version=$VERSION" -o "$DESTINATION"/share/lima/lima-guestagent.Linux-aarch64 ./cmd/lima-guestagent
	`)

	// Now, lima needs CGO is required for the build
	glg.Config.Target.GOOS = golang.GoOS(v.OS)
	glg.Config.Target.GOARCH = golang.GoArch(v.Architecture)
	glg.Config.CGO.CGO_ENABLED = 1
	glg.Config.CGO.GO_EXTLINK_ENABLED = 1

	tarch := v.Architecture

	if v == coretypes.DarwinUniversal {
		tarch = coretypes.Universal

		glg.Config.Target.GOARCH = golang.Amd64
		glg.Do(`
			cd "$LIMA_ROOT"
			go build -v -ldflags="-X "$PACKAGE"/pkg/version.Version=$VERSION" -o "$DESTINATION"/bin/limactl_amd64 ./cmd/limactl
		`)

		glg.Config.Target.GOARCH = golang.Arm64
		glg.Do(`
			cd "$LIMA_ROOT"
			go build -v -ldflags="-X "$PACKAGE"/pkg/version.Version=$VERSION" -o "$DESTINATION"/bin/limactl_arm64 ./cmd/limactl
		`)

		glg.Mount["/sign"] = &wrapllb.State{
			Source: ent,
		}

		glg.Do(`
			lipo -create -output "$DESTINATION"/bin/codecomet.upstream.limactl "$DESTINATION"/bin/limactl_arm64 "$DESTINATION"/bin/limactl_amd64
			LD_LIBRARY_PATH=/opt/macosxcross/bin codesign -f --entitlements /sign/entitlements.plist --sign - "$DESTINATION"/bin/codecomet.upstream.limactl
			rm "$DESTINATION"/bin/limactl_*
		`)
	} else {
		glg.Config.Target.GOOS = golang.GoOS(v.OS)
		glg.Config.Target.GOARCH = golang.GoArch(v.Architecture)
		glg.Do(`
			cd "$LIMA_ROOT"
			go build -v -ldflags="-X "$PACKAGE"/pkg/version.Version=$VERSION" -o "$DESTINATION"/bin/codecomet.upstream.limactl ./cmd/limactl
		`)

	}

	return outy.File(llb.Copy(glg.Destination, "/", "/"+v.OS+"/"+tarch, &llb.CopyInfo{
		CreateDestPath: true,
	}))
}

func build() llb.State {
	repo := "github.com/lima-vm/lima"
	version := "v0.14.2"

	source := codecomet.From((&codecomet.Git{
		Reference:  version,
		KeepGitDir: true,
	}).Parse(fmt.Sprintf("https://%s.git", repo)))

	outx := []llb.State{
		buildone(source, repo, version, entitl, coretypes.LinuxArm64),
		buildone(source, repo, version, entitl, coretypes.LinuxAmd64),
		buildone(source, repo, version, entitl, coretypes.DarwinUniversal),
	}

	return llb.Merge(outx)
}
