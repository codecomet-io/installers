package macos

import (
	_ "embed"
	"fmt"
	"github.com/codecomet-io/installers/sdk/base"
	"github.com/codecomet-io/installers/sdk/base/c"
	"github.com/codecomet-io/installers/sdk/base/debian"
	"github.com/codecomet-io/installers/sdk/base/llvm"
	"github.com/codecomet-io/installers/sdk/bin/apt"
	"github.com/codecomet-io/installers/sdk/bin/bash"
	"github.com/codecomet-io/installers/sdk/bin/patch"
	"github.com/codecomet-io/isovaline/sdk/codecomet"
	"github.com/codecomet-io/isovaline/sdk/wrapllb"
	"github.com/codecomet-io/isovaline/sdk/wrapllb/platform"
	"github.com/moby/buildkit/client/llb"
	"strconv"
)

//go:embed patches/sdk_version.patch
var sdk_patch string

func pack(sdkPath string, sdkVersion Version) llb.State {
	grp := &wrapllb.Group{
		ID:           fmt.Sprintf("MACOSSDK-%s-%s", sdkVersion, sdkPath),
		Name:         fmt.Sprintf("Packaging macOS SDK %s from %s", sdkVersion, sdkPath),
		DoNotDisplay: false,
	}

	// XXX wrong platform, but not a big deal here
	toolingImage := base.Debian(debian.Bullseye, platform.DefaultPlatform)

	aptGet := apt.New(toolingImage)
	aptGet.Group = grp
	aptGet.Install("bzip2")

	localSdk := codecomet.From(&codecomet.Local{
		Base: codecomet.Base{
			Group: grp,
		},
		// TODO investigate iOS
		Path: sdkPath + "/Contents/Developer/Platforms/MacOSX.platform/Developer/SDKs",
	})

	// XXX Still unclear if we really need all that
	localInclude := codecomet.From(&codecomet.Local{
		Base: codecomet.Base{
			Group: grp,
		},
		Path: sdkPath + "/Contents/Developer/Toolchains/XcodeDefault.xctoolchain/usr/include",
	})

	// XXX Still unclear if we really need all that
	localShare := codecomet.From(&codecomet.Local{
		Base: codecomet.Base{
			Group: grp,
		},
		Path: sdkPath + "/Contents/Developer/Toolchains/XcodeDefault.xctoolchain/usr/share",
	})

	bsh := bash.New(aptGet.State)
	bsh.Group = grp
	bsh.ReadOnly = true
	// Be sure to have ample space for the SDK
	bsh.TMPFSSize = 1024
	bsh.Dir = "/output"
	bsh.Mount["/output"] = &wrapllb.State{
		Source: llb.Scratch(),
	}
	bsh.Mount["/input/SDK"] = &wrapllb.State{
		ReadOnly: true,
		Source:   localSdk,
	}
	bsh.Mount["/input/include"] = &wrapllb.State{
		ReadOnly: true,
		Source:   localInclude,
	}
	bsh.Mount["/input/share"] = &wrapllb.State{
		ReadOnly: true,
		Source:   localShare,
	}

	bsh.Env["SDK_VERSION"] = string(sdkVersion)
	bsh.Env["DESTINATION"] = "/output"

	bsh.Run("Packing SDK", `
	readonly SDK_LOCATION=/input/SDK
	readonly LIBCXXDIR_LOCATION=/input/include
	readonly MANDIR_LOCATION=/input/share

	for sdk in "$SDK_LOCATION"/*; do
		# XXX how is this going to work with non macos sdks?
		[ "$SDK_VERSION" == "" ] || [ "$sdk" == "$SDK_LOCATION/MacOSX$SDK_VERSION.sdk" ] || continue
		sdkName="$(basename "$sdk")"
		temporary="${TMPDIR:-/tmp}/packing_$sdkName"
		mkdir -p "$temporary"
		cp -R "$(readlink -f "$sdk")" "$temporary/$sdkName"
		mkdir -p "$temporary"/$sdkName/usr/include
		mkdir -p "$temporary"/$sdkName/usr/share
		cp -Rf "$LIBCXXDIR_LOCATION"/* "$temporary/$sdkName"/usr/include
		cp -Rf "$MANDIR_LOCATION"/* "$temporary/$sdkName"/usr/share
		cd "$temporary" || exit
		tar -cf - * | bzip2 -5 - > "$DESTINATION/$sdkName.tar.bz2"
		cd - >/dev/null || exit
	done
	`)

	return bsh.Mount["/output"].Source
}

func buildSignTool(grp *wrapllb.Group, debianVersion debian.Version, plt *platform.Platform) llb.State {
	// Get Sigtool (MIT license)
	git := codecomet.From((&codecomet.Git{
		Reference: "c6242cb29c412168f771e97d75417e55af6cdb2e",
	}).Parse("https://github.com/thefloweringash/sigtool.git"))

	toolingImage := base.Debian(debianVersion, plt)
	toolingImage = c.Add(toolingImage, []*platform.Platform{
		plt, // platform.DefaultPlatform,
	})

	packages := []interface{}{
		"libssl-dev",
	}

	aptget := apt.New(toolingImage)
	aptget.Group = grp
	aptget.Install(packages...)

	bsh := bash.New(aptget.State)
	bsh.Group = grp
	bsh.ReadOnly = true
	bsh.Dir = "/codecomet"

	bsh.Mount["/input"] = &wrapllb.State{
		NoOutput: true,
		Source:   git,
	}

	bsh.Mount["/opt/macosxcross/bin"] = &wrapllb.State{
		Source: llb.Scratch().File(llb.Mkdir("/opt/macosxcross/bin", 0755, llb.WithParents(true))),
		Path:   "/opt/macosxcross/bin",
	}

	bsh.Run("Building signtool", `
		cmake /input
		make
		# XXX hold on your hat cowboy
		cp sigtool /opt/macosxcross/bin
		cp codesign /opt/macosxcross/bin
		cp libsigtool.so /opt/macosxcross/bin
		cd /opt/macosxcross/bin
		ln -s x86_64-apple-darwin22.2-codesign_allocate codesign_allocate
	`)

	return bsh.Mount["/opt/macosxcross/bin"].Source
}

// Need a C state
func Add(sdkPath string, sdkVersion Version, debianVersion debian.Version, llvmVersion llvm.Version, plt *platform.Platform) llb.State {
	grp := &wrapllb.Group{
		ID:   fmt.Sprintf("MACOSTOOLCHAIN-%s-%s-%s-%s", sdkVersion, sdkPath, debianVersion, llvmVersion),
		Name: fmt.Sprintf("Compiling macOS toolchain %s from %s", sdkVersion, sdkPath),
	}

	sdkState := pack(sdkPath, sdkVersion)

	// Get the source
	// GPL 2 license
	git := codecomet.From((&codecomet.Git{
		Reference: osxCrossVersion,
	}).Parse("https://github.com/tpoechtrager/osxcross.git"))
	// Patch it
	git = patch.Patch(git, sdk_patch)

	// Get or build a builder image
	// The get version
	// toolingImage := base.C(debianVersion, llvmVersion, false, plt)
	// The build version
	toolingImage := base.Debian(debianVersion, plt)
	toolingImage = c.Add(toolingImage, []*platform.Platform{
		// XXXXXX this is... concerning...
		// platform.DefaultPlatform,
		plt,
	})
	toolingImage = llvm.Add(toolingImage, debianVersion, llvmVersion)

	// toolingImage = InstallBuildEssential(toolingImage)
	packages := []interface{}{
		// Default build essential
		// "make",
		// "cmake",
		"bzip2",
		"git",
		"python3",
		"libbz2-dev",
		"zlib1g-dev",
		"libssl-dev",
		"libxml2-dev",
		"lzma-dev",
		"uuid-dev",
	}
	// From repo:
	//  clang llvm-dev libxml2-dev uuid-dev \
	//  libssl-dev bash patch make tar xz-utils bzip2 gzip sed cpio libbz2-dev \
	//  zlib1g-dev

	aptget := apt.New(toolingImage)
	aptget.Group = grp
	aptget.Install(packages...)

	bsh := bash.New(aptget.State)
	bsh.Group = grp
	// Because we need to add symlinks to /usr/bin/clang, as macoscross is confused
	bsh.ReadOnly = false
	// bsh.Dir = "/opt/macosxcross"
	bsh.Dir = "/codecomet"

	bsh.Mount["/input"] = &wrapllb.State{
		NoOutput: true,
		Source:   git,
	}

	bsh.Mount["/input/tarballs"] = &wrapllb.State{
		ReadOnly: true,
		Source:   sdkState,
	}

	bsh.Mount["/opt/macosxcross"] = &wrapllb.State{
		Source: llb.Scratch().File(llb.Mkdir("/opt/macosxcross", 0755, llb.WithParents(true))),
		Path:   "/opt/macosxcross",
	}

	bsh.Env["SOURCE_DIR"] = "/input"
	bsh.Env["TARGET_DIR"] = "/opt/macosxcross"
	bsh.Env["SDK_VERSION"] = string(sdkVersion)
	bsh.Env["OSX_VERSION_MIN"] = "10.9"
	bsh.Env["UNATTENDED"] = "1"
	bsh.Env["OCDEBUG"] = "" //"1"
	bsh.Env["LLVMVERSION"] = strconv.Itoa(int(llvmVersion))
	bsh.Run("Building macOS clang cross compilation toolchain", `
		# Using llvm17 means we have to set these to workaround osxcross issues
		export CC="clang-$LLVMVERSION"
		export CXX="clang++-$LLVMVERSION"
		ln -s /usr/lib/llvm-$LLVMVERSION/bin/clang /usr/bin/clang
		ln -s /usr/lib/llvm-$LLVMVERSION/bin/clang++ /usr/bin/clang++
	
		export INSTALLPREFIX="$TARGET_DIR"

		"$SOURCE_DIR"/build.sh
	`)

	codesign := buildSignTool(grp, debianVersion, plt)

	return llb.Merge([]llb.State{bsh.Mount["/opt/macosxcross"].Source, codesign})
}

// "LDFLAGS":    "",
// XXX -fcf-protection will fail the build
// "CFLAGS":   "-Werror=implicit-function-declaration -Werror=format-security -Wall -O3 -grecord-gcc-switches -g -Wp,-D_GLIBCXX_ASSERTION -D_FORTIFY_SOURCE=2 -pipe -fexceptions -fstack-protector-strong",
// "CXXFLAGS": "-Werror=implicit-function-declaration -Werror=format-security -Wall -O3 -grecord-gcc-switches -g -Wp,-D_GLIBCXX_ASSERTION -D_FORTIFY_SOURCE=2 -pipe -fexceptions -fstack-protector-strong",

/*
				tapigit := codecomet.From((&codecomet.Git{
					Reference: "1100.0.11",
				}).Parse("https://github.com/tpoechtrager/apple-libtapi.git"))

				bsh.Run("Building macOS clangx: tapi", `
	cd "apple-libtapi"

	export CC="clang-17"
	export CXX="clang++-17"

	export TAPI_REPOSITORY=1100.0.11
	export TAPI_VERSION=11.0.0 # ?!

	INCLUDE_FIX="-I $(pwd)/../src/llvm/projects/clang/include "
	INCLUDE_FIX+="-I $(pwd)/projects/clang/include "

	cmake ../src/llvm \
	 -DCMAKE_CXX_FLAGS="$INCLUDE_FIX" \
	 -DLLVM_INCLUDE_TESTS=OFF \
	 -DCMAKE_BUILD_TYPE=RELEASE \
	 -DCMAKE_INSTALL_PREFIX=$TARGET_DIR \
	 -DTAPI_REPOSITORY_STRING=$TAPI_REPOSITORY \
	 -DTAPI_FULL_VERSION=$TAPI_VERSION

	make clangBasic -j$(nproc)
	make libtapi -j$(nproc)
	make install-libtapi install-tapi-headers -j$(nproc)
	`)

*/
