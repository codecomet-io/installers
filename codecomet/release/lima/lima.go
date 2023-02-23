package main

import (
	_ "embed"
	"fmt"
	golang2 "github.com/codecomet-io/go-sdk/base/golang"
	"github.com/codecomet-io/go-sdk/bin/golang"
	"github.com/codecomet-io/go-sdk/codecomet"
	"github.com/codecomet-io/go-sdk/codecomet/wrap"
	"github.com/codecomet-io/go-sdk/controller"
	"github.com/codecomet-io/go-sdk/execcontext/debian"
	"github.com/codecomet-io/go-sdk/fileset"
	"github.com/moby/buildkit/client/llb"
)

//go:embed templates/codecomet-bullseye.yaml
var vmtpl string

//go:embed entitlements/lima.plist
var entitl string

//go:embed patches/share_path.patch
var share_patch string

func main() {
	controller.Init()

	// XXX this is largely incorrect - these are GoLang platforms, not OCI
	// XXX Geez hard coded host path dammmmmm
	outx := build()

	controller.Get().Exporter = &controller.Export{
		Local: "release/mark-I", // os.Args[2],
		// Oci: "oci-tester/exp.tar",
	}

	controller.Get().Do(outx)
}

func buildone(src codecomet.FileSet, repo string, version string, entitlements string, v *codecomet.Platform) llb.State {

	ent := fileset.File("entitlements.plist", []byte(entitlements), 0400)

	destination := fileset.New()
	destination.MkDir("/bin", &codecomet.MkDirOptions{
		Mode: 0700,
	})
	destination.MkDir("/share/codecomet/examples", &codecomet.MkDirOptions{
		Mode: 0700,
	})
	destination.AddFile("/share/codecomet/examples/codecomet-bullseye.yaml", []byte(vmtpl), &codecomet.AddFileOptions{
		Mode: 0600,
	})

	glg := golang.New(debian.Bullseye, golang2.Go1_19, true, true)
	glg.Source = src.GetInternalState()
	glg.Destination = destination.GetInternalState()

	// These are mere shell wrapping a call to limactl
	// symlinking TBD

	glg.Env["DESTINATION"] = "/output"
	glg.Env["LIMA_ROOT"] = "/input"
	glg.Env["PACKAGE"] = repo
	glg.Env["VERSION"] = version

	// XXX should this be in go instead?
	glg.Env["MACOSX_DEPLOYMENT_TARGET"] = "13.0"

	glg.Do(`
		# ln -s codecomet-bullseye.yaml "$DESTINATION"/share/lima/examples/default.yaml
		ln -s codecomet-bullseye.yaml "$DESTINATION"/share/codecomet/examples/default.yaml
	`)

	glg.Config.Target.GOOS = golang.Linux
	glg.Config.Target.GOARCH = golang.Amd64

	glg.Do(`
		go build -ldflags="-s -w -X "$PACKAGE"/pkg/version.Version=$VERSION" -o "$DESTINATION"/share/codecomet/lima-guestagent.Linux-x86_64 ./cmd/lima-guestagent
	`)

	glg.Config.Target.GOARCH = golang.Arm64
	glg.Do(`
		go build -ldflags="-s -w -X "$PACKAGE"/pkg/version.Version=$VERSION" -o "$DESTINATION"/share/codecomet/lima-guestagent.Linux-aarch64 ./cmd/lima-guestagent
	`)

	// Now, lima needs CGO is required for the build
	glg.Config.Target.GOOS = golang.GoOS(v.OS)
	glg.Config.Target.GOARCH = golang.GoArch(v.Architecture)
	glg.Config.CGO.CGO_ENABLED = 1
	glg.Config.CGO.GO_EXTLINK_ENABLED = 1

	tarch := v.Architecture

	if v == codecomet.DarwinUniversal {
		tarch = codecomet.ArchUniversal

		glg.Config.Target.GOARCH = golang.Amd64
		glg.Do(`
			go build -v -ldflags="-X "$PACKAGE"/pkg/version.Version=$VERSION" -o "$DESTINATION"/bin/limactl_amd64 ./cmd/limactl
		`)

		glg.Config.Target.GOARCH = golang.Arm64
		glg.Do(`
			go build -v -ldflags="-X "$PACKAGE"/pkg/version.Version=$VERSION" -o "$DESTINATION"/bin/limactl_arm64 ./cmd/limactl
		`)

		glg.Mount["/sign"] = &wrap.State{
			Source: ent.GetInternalState(),
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
			go build -v -ldflags="-X "$PACKAGE"/pkg/version.Version=$VERSION" -o "$DESTINATION"/bin/codecomet.upstream.limactl ./cmd/limactl
		`)
	}

	return fileset.Copy(fileset.New().Adopt(glg.Destination), "/", fileset.New(), "/"+v.OS+"/"+tarch, &codecomet.CopyOptions{}).GetInternalState()

}

func build() llb.State {
	repo := "github.com/lima-vm/lima"
	version := "v0.14.2"

	gits := fileset.Git(fmt.Sprintf("https://%s.git", repo))
	gits.Reference = version
	gits.KeepGitDir = true
	gits.Patch([]string{share_patch}, &codecomet.PatchOptions{})

	return fileset.Merge([]codecomet.FileSet{
		fileset.New().Adopt(buildone(gits, repo, version, entitl, codecomet.LinuxArm64)),
		fileset.New().Adopt(buildone(gits, repo, version, entitl, codecomet.LinuxAmd64)),
		fileset.New().Adopt(buildone(gits, repo, version, entitl, codecomet.DarwinUniversal)),
	}, &codecomet.MergeOptions{}).GetInternalState()

}
