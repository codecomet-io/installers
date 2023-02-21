package main

import (
	"github.com/codecomet-io/go-sdk/codecomet"
	"github.com/codecomet-io/go-sdk/codecomet/action"
	"github.com/codecomet-io/go-sdk/codecomet/core"
	"github.com/codecomet-io/go-sdk/codecomet/interfaces"
	"github.com/codecomet-io/go-sdk/codecomet/workspace"
	"github.com/codecomet-io/go-sdk/controller"
	"github.com/codecomet-io/go-sdk/fileset"
	"github.com/codecomet-io/go-sdk/folder"
)

func main() {
	controller.Init()

	// git := source.Git("https://github.com/thefloweringash/sigtool.git#c6242cb29c412168f771e97d75417e55af6cdb2e")

	myPipeline := codecomet.Pipeline{}

	outputofAction1 := myPipeline.Run(&action.Action{
		Name:    "Going to build X from go-sdk repository",
		Input:   fileset.Git("github.com/codecomet-io/go-sdk"),
		Env:     map[string]string{},
		Command: []string{"go build foo"},
	})

	outputofAction2 := myPipeline.Run(&action.Action{
		Name:    "Going to build Y from go-sdk repository",
		Input:   fileset.Git("github.com/codecomet-io/go-sdk"),
		Env:     map[string]string{},
		Command: []string{"yarn install"},
	})

	outputofAction2 := myPipeline.Run(&action.BradDeployLimitedActiononEC2{
		Input: outputOfOtherAction,
		Secret: []string{"xyz"},
		Region: "us-east-1",
		InstanceType: "t2.micro",
	})



		// Utility Actions
	outputAction1.Merge(outputofAction2)

	folderX.AddFile("/myfile", "lalalalalla")
	folderX.MkDir("/mydirectory")
	folderX.Symlink("/mylinktarget", "/mylinklocation")
	folderX.Patch("lalalala")

	output := codecomet.Merge(outputofAction1, outputofAction2)
	codecomet.Copy(outputofAction1, "/foo", outputofAction2, "/bar", &codecomet.CopyOptions{})


	myPipeline.Export({
		// Tarball: "result.tar"
		DockerImage: "docker.io/brad/foo"
	})

	/*
		git := source.Git("github.com/mgklf")
		http := source.HTTP("https://gsdglkdsf")

		http.Patch([]string{"patchie patch"})
		http.AddFile("rajiv", "content")
		http.MkDir("rajiv")

		newSource := fileset.Merge([]codecomet.FileSet{git, loc, http}, &codecomet.Nameable{})
		fileset.Copy(git, "/cmd", loc, "/fromgit", &codecomet.CopyOptions{})
	*/

	// container := source.Image("docker.io/library/debian:bullseye-slim")

	// im := codecomet.internalToolImage().GetInternalState()

	// Sequential by default, which is governed by the output
	first.Run(&action.Action{
		Name:    "test 2",
		Command: []string{"bash", "-c", "--", "cat /codecomet/input/insideinput > /codecomet/output/frominsideinputtooutput"},
	})

	// result := first.Output

	second := codecomet.Pipeline{
		Input: fileset.Merge([]codecomet.FileSet{first.Output}, &codecomet.Nameable{}),
	}

	second.Run(&action.Action{
		Name: "test parallel",
		Env: map[string]string{
			"foo": "bar",
		},
		Command: []string{"bash", "-c", "--", "cat /codecomet/input/insideinput > /codecomet/output/portedfromotherpipeinput"},
	})

	result := fileset.Merge([]codecomet.FileSet{first.Output, second.Output}, &codecomet.Nameable{})

	// codecomet.ExportToTarball(result)
	// codecomet.ExportToDockerImage("docker.io/codecometio/testfo", result)

	outx := result.GetInternalState()

	/*
		na := codecomet.NGNewPipe()
		// na.Use(source.Image("docker.io/library/debian:bullseye-slim"))
		na.Dir = "/tmp"
		na.Hostname = "codecomet-foofoo"

		na.AddCache(&codecomet.Cache{UniqueDescription: "Llalala", CacheSharingMode: codecomet.CacheModeShared}, "/cacher")

		na.AddTemp(&codecomet.NGTemp{Size: 1024}, "/temper")

		na.Mount(&codecomet.Scratch{}, "/mounter", &codecomet.MountOptions{  })

		na.Env["foo"] = "bar"

	*/

	/*
		na.Run(&codecomet.Action{
			Name: "foofoo",
			Args: []string{
				"sh", "-c", "--", `
					touch /foo
					touch /mounter/mountfoo2 || {
						echo "mount fail"
					}
					touch /temper/tmpfsfoo || {
						echo "temp fail"
					}
					touch /cacher/cachefoo || {
						echo "cache fail"
					}
					env
					`,
			}})

		na.Run(&codecomet.Action{
			Name: "barbar",
			Args: []string{
				"sh", "-c", "--", `
					ls -lA /mounter || true
					ls -lA /temper || true
					ls -lA /cacher || true
					env
					exit 123
					`,
			}})

		na.Run(&codecomet.NGBashRunner{
			Args: []string{"echo lol"},
		})

		na.Run(&codecomet.NGBashRunner{
			Args: []string{"echo bar; exit 123"},
		})

	*/

	// outx := na.GetRoot().GetInternalState()
	controller.Get().Exporter = &controller.Export{
		Local: "debug-output", // os.Args[2],
	}

	controller.Get().Do(outx)

	// na.GetMount()

	// var opts []llb.ImageOption

	// XXX this is somewhat bad - host platform is not the same thing as builder platform
	// opts = append(opts, llb.Platform(*codecomet.DefaultPlatform))

	// Always force pull, actually. Might be useful not to in certain conditions, but use case must be clarified first
	// opts = append(opts, llb.ResolveModeForcePull)

	/*
		outx := fileset.Merge([]codecomet.FileSet{
			git,
			loc,
		}, &core.ActionOptions{}).GetInternalState()

		img := &codecomet.Image{
			Registry: "docker.io",
			Owner:    "codecometio",
			Name:     "distro_debian",
			Tag:      "bullseye-arm64",
			Platform: codecomet.DefaultPlatform,
		}

		img.ActionName = "foo"
		img.ActionGroup = codecomet.InternalGroup

		outx := img.GetInternalState()

		deb := root.Debian(debian.Bullseye, platform.DefaultPlatform).GetInternalState()
		outx = deb

	*/

	/*
	 */

	// basePython := root.Python(debian.Bullseye, false, platform.DefaultPlatform).GetInternalState()
	/*
		myLocalProject := codecomet.Local{
			Path: "xxxx",
		}

	*/
	/*
		bsh := bash.New(basePython)
		//bsh.Mount["/source"] = &wrap.State{
		//	Source: myLocalProject.GetInternalState(),
		//}
		bsh.Dir = "/"

		bsh.Run("pytest foo")

		outx := bsh.State

	*/

	/*
		"nonexistent"
		#2 ERROR: repository does not contain ref nonexistent, output: ""

		"c06d0e5787308a4c10ebd0c4e583127e3de380dd"
		#1 16.78 fatal: reference is not a tree: c06d0e5787308a4c10ebd0c4e583127e3de380dd
		#1 ERROR: failed to checkout remote https://github.com/codecomet-io/installers.git: exit status 128



			{
			url:      "http://github.com/moby/buildkit",
			protocol: HTTPProtocol,
			remote:   "github.com/moby/buildkit",
		},
		{
			url:      "https://github.com/moby/buildkit",
			protocol: HTTPSProtocol,
			remote:   "github.com/moby/buildkit",
		},
		{
			url:      "git@github.com:moby/buildkit.git",
			protocol: SSHProtocol,
			remote:   "github.com:moby/buildkit.git",
		},
		{
			url:      "nonstandarduser@example.com:/srv/repos/weird/project.git",
			protocol: SSHProtocol,
			remote:   "example.com:/srv/repos/weird/project.git",
		},
		{
			url:      "ssh://root@subdomain.example.hostname:2222/root/my/really/weird/path/foo.git",
			protocol: SSHProtocol,
			remote:   "subdomain.example.hostname:2222/root/my/really/weird/path/foo.git",
		},
		{
			url:      "git://host.xz:1234/path/to/repo.git",


	*/

	// Non existent commit ref
	// displays: fatal: Not a valid object name 0937cc6e1ea14bcd6d5ec79bf4035d5a088b6174^{commit}
	// Fails properly somewhat
	// Hits it once
	// #5 39.37 Initialized empty Git repository in /var/lib/buildkit/runc-overlayfs/snapshots/snapshots/6/fs/.git/
	// fatal: git cat-file: could not get object info
	// g := llb.Git("ssh://git@github.com/codecomet-io/cli.git", "0937cc6e1ea14bcd6d5ec79bf4035d5a088b6174", llb.KeepGitDir())
	// g := llb.Git("git@github.com:codecomet-io/cli.git", "0937cc6e1ea14bcd6d5ec79bf4035d5a088b6174", llb.KeepGitDir())

	// Random non existing ref
	// Fails the build
	// Hits it once, less noise
	// g := llb.Git("ssh://git@github.com/codecomet-io/cli.git", "foo", llb.KeepGitDir())
	// g := llb.Git("git@github.com:codecomet-io/cli.git", "foo", llb.KeepGitDir())

	// Existing commit ref
	// Works
	// Hits it once
	// Displays wrongly:
	// #4 15.90 fatal: Not a valid object name ad41d95d0f171f6b8908b7d1aedcc3b5f83d11c2^{commit}
	// g := llb.Git("ssh://git@github.com/codecomet-io/cli.git", "ad41d95d0f171f6b8908b7d1aedcc3b5f83d11c2", llb.KeepGitDir())
	// g := llb.Git("git@github.com:codecomet-io/cli.git", "ad41d95d0f171f6b8908b7d1aedcc3b5f83d11c2", llb.KeepGitDir())

	// Valid branch
	// Works
	// Hits it twice
	// g := llb.Git("ssh://git@github.com/codecomet-io/cli.git", "acme-attestation", llb.KeepGitDir())
	// g := llb.Git("git@github.com:codecomet-io/cli.git", "acme-attestation", llb.KeepGitDir())

	// Valid tag
	// Works
	// Hits it twice
	// g := llb.Git("ssh://git@github.com/codecomet-io/cli.git", "v0.21.0", llb.KeepGitDir())
	// g := llb.Git("git@github.com:codecomet-io/cli.git", "v0.21.0", llb.KeepGitDir())

	// Works
	// Hits things 3 / 4? times
	// g := llb.Git("ssh://git@github.com/codecomet-io/cli.git", "", llb.KeepGitDir())
	// g := llb.Git("git@github.com:codecomet-io/cli.git", "", llb.KeepGitDir())
	// g := llb.Git("https://github.com/codecomet-io/cli.git", "", llb.KeepGitDir())

	/*
		outx := codecomet.Execute(root.Debian(debian.Bullseye, codecomet.DefaultPlatform).GetInternalState(), &legacy.Exec{

			Mount: map[string]*wrap.State{
				"/g": &wrap.State{
					Source: g,
				},
			},
			// And the rest
			Args: []string{"bash", "-c", "--", `
				apt-get update -qq
				apt-get install git -qq
				cd /g
				git log -1
			`},
		})

	*/

	/*

		bsh := bash.New(root.Debian(debian.Bullseye, platform.DefaultPlatform).GetInternalState())
		bsh.CanFail = true
		bsh.Mount["/source"] = &wrap.State{
			Source: *git.GetInternalState(),
		}
		bsh.Mount["/local"] = &wrap.State{
			Source: *loc.GetInternalState(),
		}
		bsh.Run("Hello world", `
			ls -lA /source
			echo "----"
			ls -lA /local
		`)

		outx := bsh.State

	*/

	/*
			git := (&codecomet.Git{
				Reference: "c6242cb29c412168f771e97d75417e55af6cdb2e",
			}).Parse("https://github.com/thefloweringash/sigtool.git").GetInternalState()

			bsh := bash.New(root.Debian(debian.Bullseye, platform.DefaultPlatform).GetInternalState())
			bsh.CanFail = true
			bsh.Mount["/source"] = &wrap.State{
				Source: git,
			}
			bsh.Run("Hello world", `
				echo "Hello I am Rajiv"
				ls -lA /nonexistent
		`)

			outx := bsh.State
	*/

}

/*
node := codecomet.New("Pipeline ID", "Pipeline Name", ...)

node2 := node.Clone()
node3 := node.Clone()

node := fileset.Merge(node, node2, node3)
// node.Merge(node2, node3)

node.Mount("/git", source.Git)
node.Mount("/http", source.Http)
node.Mount("/local", source.Local)
node.Unmount("/git")
// Implied
// Reset every action
node.Mount("/tmp", &TMPFS{})
// Carry over, but may be changed concurrently
node.Mount("/share", &Share{Mode: Shared})
// Carry over, unless we fork
node.Mount("/output", &Scratch{})


//plan.WithDebian()
//plan.Use(&Debian{})

action := debian.Attach(node, debian.Bullseye)

action := WithDebian(node, debian.Bullseye)

action.AptGet
action.Go
action.Python
action.Node
// All embed Bash
action.Bash
// Bash gives
.Dir
.Env
.Group
.Name
.Output
action.Output <- get the resulting folder

// plan.Node = debian.New()

Copy

*/
