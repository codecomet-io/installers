package golang

import "strconv"

type OnOffAuto string

const (
	On   OnOffAuto = "on"
	Off  OnOffAuto = "off"
	Auto OnOffAuto = "auto"
)

type DisabledEnabled uint

const (
	Disabled DisabledEnabled = iota
	Enabled
)

type Config struct {
	// https://golang.org/ref/mod#mod-commands
	GOARCH   GoArch
	GOOS     GoOS
	GOARM    GoARM
	GO386    Go386
	GOAMD64  GoAMD64
	GOMIPS   GoMips
	GOMIPS64 GoMips64
	GOPPC64  GoPPC64
	GOWASM   GoWasm

	CGO_ENABLED DisabledEnabled

	//
	GO111MODULE OnOffAuto
	// Root of go folder (/source)
	GOPATH string
	// SDK root
	GOROOT string
}

func (c *Config) ToEnv() map[string]string {
	envs := make(map[string]string)
	envs["GOARCH"] = string(c.GOARCH)
	envs["GOOS"] = string(c.GOOS)
	envs["GOARM"] = string(c.GOARM)
	envs["GO386"] = string(c.GO386)
	envs["GOAMD64"] = string(c.GOAMD64)
	envs["GOMIPS"] = string(c.GOMIPS)
	envs["GOMIPS64"] = string(c.GOMIPS64)
	envs["GOPPC64"] = string(c.GOPPC64)
	envs["GOWASM"] = string(c.GOWASM)

	envs["CGO_ENABLED"] = strconv.Itoa(int(c.CGO_ENABLED))

	envs["GO111MODULE"] = string(c.GO111MODULE)
	envs["GOPATH"] = c.GOPATH
	envs["GOROOT"] = c.GOROOT

	return envs
}

const (
	// C/C++/CGO stuff
	// https://news.ycombinator.com/item?id=18874113
	// https://developers.redhat.com/blog/2018/03/21/compiler-and-linker-flags-gcc
	// https://gcc.gnu.org/onlinedocs/gcc/Warning-Options.html
	WARNING_OPTIONS = "-Werror=implicit-function-declaration -Werror=format-security -Wall"
	// https://gcc.gnu.org/onlinedocs/gcc/Optimize-Options.html#Optimize-Options
	OPTIMIZATION_OPTIONS = "-O3"
	// https://gcc.gnu.org/onlinedocs/gcc/Debugging-Options.html#Debugging-Options
	DEBUGGING_OPTIONS = "-grecord-gcc-switches -g"
	// https://gcc.gnu.org/onlinedocs/gcc/Preprocessor-Options.html#Preprocessor-Options
	PREPROCESSOR_OPTIONS = "-Wp,-D_GLIBCXX_ASSERTION -D_FORTIFY_SOURCE=2"
	// https://gcc.gnu.org/onlinedocs/gcc/Instrumentation-Options.html
	COMPILER_OPTIONS = "-pipe -fexceptions -fstack-protector-strong"
	// Only linux
	COMPILER_OPTIONS_LINUX = "-fstack-clash-protection"
	// Only linux amd64
	COMPILER_OPTIONS_LINUX_AMD64 = "-mcet -fcf-protection"
	// Only Darwin
	COMPILER_OPTIONS_DARWIN = "-fcf-protection"

	// XXX -s?
	// Aggregate all the options
	CFLAGS = WARNING_OPTIONS + " " + OPTIMIZATION_OPTIONS + " " + DEBUGGING_OPTIONS + " " + PREPROCESSOR_OPTIONS + " " + COMPILER_OPTIONS + " -s"
	// Note: Werror=implicit-function-declaration is not allowed for CXX
	CXXFLAGS = "-Werror=format-security -Wall " + OPTIMIZATION_OPTIONS + " " + DEBUGGING_OPTIONS + " " + PREPROCESSOR_OPTIONS + " " + COMPILER_OPTIONS + " -s"
)

/*

GCCGO
The gccgo command to run for 'go build -compiler=gccgo'.
GOBIN
The directory where 'go install' will install a command.
GODEBUG
Enable various debugging facilities. See 'go doc runtime'
for details.
GOFLAGS
A space-separated list of -flag=value settings to apply
to go commands by default, when the given flag is known by
the current command. Each entry must be a standalone flag.
Because the entries are space-separated, flag values must
not contain spaces. Flags listed on the command line
are applied after this list and therefore override it.
GOINSECURE
Comma-separated list of glob patterns (in the syntax of Go's path.Match)
of module path prefixes that should always be fetched in an insecure
manner. Only applies to dependencies that are being fetched directly.
GOINSECURE does not disable checksum database validation. GOPRIVATE or
GONOSUMDB may be used to achieve that.

GOPROXY
URL of Go module proxy. See https://golang.org/ref/mod#environment-variables
and https://golang.org/ref/mod#module-proxy for details.
GOPRIVATE, GONOPROXY, GONOSUMDB
Comma-separated list of glob patterns (in the syntax of Go's path.Match)
of module path prefixes that should always be fetched directly
or that should not be compared against the checksum database.
See https://golang.org/ref/mod#private-modules.
GOSUMDB
The name of checksum database to use and optionally its public key and
URL. See https://golang.org/ref/mod#authenticating.
GOTMPDIR
The directory where the go command will write
temporary source files, packages, and binaries.
GOVCS
Lists version control commands that may be used with matching servers.
See 'go help vcs'.
GOWORK
In module aware mode, use the given go.work file as a workspace file.
By default or when GOWORK is "auto", the go command searches for a
file named go.work in the current directory and then containing directories
until one is found. If a valid go.work file is found, the modules
specified will collectively be used as the main modules. If GOWORK
is "off", or a go.work file is not found in "auto" mode, workspace
mode is disabled.

Environment variables for use with cgo:

AR
The command to use to manipulate library archives when
building with the gccgo compiler.
The default is 'ar'.
CC
The command to use to compile C code.
CGO_ENABLED
Whether the cgo command is supported. Either 0 or 1.
CGO_CFLAGS
Flags that cgo will pass to the compiler when compiling
C code.
CGO_CFLAGS_ALLOW
A regular expression specifying additional flags to allow
to appear in #cgo CFLAGS source code directives.
Does not apply to the CGO_CFLAGS environment variable.
CGO_CFLAGS_DISALLOW
A regular expression specifying flags that must be disallowed
from appearing in #cgo CFLAGS source code directives.
Does not apply to the CGO_CFLAGS environment variable.
CGO_CPPFLAGS, CGO_CPPFLAGS_ALLOW, CGO_CPPFLAGS_DISALLOW
Like CGO_CFLAGS, CGO_CFLAGS_ALLOW, and CGO_CFLAGS_DISALLOW,
but for the C preprocessor.
CGO_CXXFLAGS, CGO_CXXFLAGS_ALLOW, CGO_CXXFLAGS_DISALLOW
Like CGO_CFLAGS, CGO_CFLAGS_ALLOW, and CGO_CFLAGS_DISALLOW,
but for the C++ compiler.
CGO_FFLAGS, CGO_FFLAGS_ALLOW, CGO_FFLAGS_DISALLOW
Like CGO_CFLAGS, CGO_CFLAGS_ALLOW, and CGO_CFLAGS_DISALLOW,
but for the Fortran compiler.
CGO_LDFLAGS, CGO_LDFLAGS_ALLOW, CGO_LDFLAGS_DISALLOW
Like CGO_CFLAGS, CGO_CFLAGS_ALLOW, and CGO_CFLAGS_DISALLOW,
but for the linker.
CXX
The command to use to compile C++ code.
FC
The command to use to compile Fortran code.
PKG_CONFIG
Path to pkg-config tool.


Special-purpose environment variables:

GCCGOTOOLDIR
If set, where to find gccgo tools, such as cgo.
The default is based on how gccgo was configured.
GOEXPERIMENT
Comma-separated list of toolchain experiments to enable or disable.
The list of available experiments may change arbitrarily over time.
See src/internal/goexperiment/flags.go for currently valid values.
Warning: This variable is provided for the development and testing
of the Go toolchain itself. Use beyond that purpose is unsupported.
GOROOT_FINAL
The root of the installed Go tree, when it is
installed in a location other than where it is built.
File names in stack traces are rewritten from GOROOT to
GOROOT_FINAL.
GO_EXTLINK_ENABLED
Whether the linker should use external linking mode
when using -linkmode=auto with code that uses cgo.
Set to 0 to disable external linking mode, 1 to enable it.
GIT_ALLOW_PROTOCOL
Defined by Git. A colon-separated list of schemes that are allowed
to be used with git fetch/clone. If set, any scheme not explicitly
mentioned will be considered insecure by 'go get'.
Because the variable is defined by Git, the default value cannot
be set using 'go env -w'.

Additional information available from 'go env' but not read from the environment:

GOEXE
The executable file name suffix (".exe" on Windows, "" on other systems).
GOGCCFLAGS
A space-separated list of arguments supplied to the CC command.
GOHOSTARCH
The architecture (GOARCH) of the Go toolchain binaries.
GOHOSTOS
The operating system (GOOS) of the Go toolchain binaries.
GOMOD
The absolute path to the go.mod of the main module.
If module-aware mode is enabled, but there is no go.mod, GOMOD will be
os.DevNull ("/dev/null" on Unix-like systems, "NUL" on Windows).
If module-aware mode is disabled, GOMOD will be the empty string.
GOTOOLDIR
The directory where the go tools (compile, cover, doc, etc...) are installed.
GOVERSION
The version of the installed Go tree, as reported by runtime.Version.

*/
