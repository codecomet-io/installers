package golang

import "fmt"

type GoOS string
type GoArch string

type GoARM string
type Go386 string
type GoAMD64 string
type GoMips string
type GoMips64 GoMips
type GoPPC64 string
type GoWasm string

const (
	VariantArm5          GoARM   = "5"
	VariantArm6          GoARM   = "6"
	VariantArm7          GoARM   = "7"
	Variant386SSE2       Go386   = "sse2"
	Variant386Softfloat  Go386   = "softfloat"
	VariantAmd64v1       GoAMD64 = "v1"
	VariantAmd64v2       GoAMD64 = "v2"
	VariantAmd64v3       GoAMD64 = "v3"
	VariantAmd64v4       GoAMD64 = "v4"
	VariantMipsHardfloat GoMips  = "hardfloat"
	VariantMipsSoftfloat GoMips  = "softfloat"
	VariantPpc64Power8   GoPPC64 = "power8"
	VariantPpc64Power9   GoPPC64 = "power9"
	VariantWasmSatconv   GoWasm  = "satconv"
	VariantWasmSignext   GoWasm  = "signext"
	VariantWasmSatSig    GoWasm  = "signext,satconv"
)

const (
	AIX       GoOS = "aix"
	Android   GoOS = "android"
	Darwin    GoOS = "darwin"
	DragonFly GoOS = "dragonfly"
	FreeBSD   GoOS = "freebsd"
	Illumos   GoOS = "illumos"
	IOS       GoOS = "ios"
	JS        GoOS = "js"
	Linux     GoOS = "linux"
	NetBSD    GoOS = "netbsd"
	OpenBSD   GoOS = "openbsd"
	Plan9     GoOS = "plan9"
	Solaris   GoOS = "solaris"
	Windows   GoOS = "windows"

	Amd64    GoArch = "amd64"
	I386     GoArch = "386"
	Arm      GoArch = "arm"
	Arm64    GoArch = "arm64"
	Ppc64le  GoArch = "ppc64le"
	Ppc64    GoArch = "ppc64"
	Mips64le GoArch = "mips64le"
	Mips64   GoArch = "mips64"
	Mipsle   GoArch = "mipsle"
	Mips     GoArch = "mips"
	S390x    GoArch = "s390x"
	Wasm     GoArch = "wasm"
)

/*
hurd

linux	loong64
linux	riscv64
*/

func osArchCheck(os GoOS, arch GoArch) error {
	switch os {
	case AIX, Android, Darwin, DragonFly, FreeBSD, Illumos, IOS, JS, Linux, NetBSD, OpenBSD, Plan9, Solaris, Windows:
		switch arch {
		case Amd64, I386, Arm, Arm64, Ppc64le, Ppc64, Mips64le, Mips64, Mipsle, Mips, S390x, Wasm:
		default:
			return fmt.Errorf("unknown architecture %q", arch)
		}
	default:
		return fmt.Errorf("unknown OS %q", os)
	}
	if os == AIX && arch != Ppc64 {
		return fmt.Errorf("AIX only supports ppc64")
	}
	if os == Android && arch != I386 && arch != Amd64 && arch != Arm && arch != Arm64 {
		return fmt.Errorf("Android only supports 386, amd64, arm and arm64")
	}
	if os == Darwin && arch != Amd64 && arch != Arm64 {
		return fmt.Errorf("Darwin only supports amd64 and arm64")
	}
	if os == DragonFly && arch != Amd64 {
		return fmt.Errorf("DragonFly only supports amd64")
	}
	if os == FreeBSD && arch != Amd64 && arch != I386 && arch != Arm {
		return fmt.Errorf("FreeBSD only supports 386, amd64, and arm")
	}
	if os == Illumos && arch != Amd64 {
		return fmt.Errorf("Illumos only supports amd64")
	}
	if os == IOS && arch != Arm64 {
		return fmt.Errorf("iOS only supports arm64")
	}
	if os == JS && arch != Wasm {
		return fmt.Errorf("JS only supports wasm")
	}
	if os == NetBSD && arch != Amd64 && arch != I386 && arch != Arm {
		return fmt.Errorf("NetBSD only supports 386, amd64, and arm")
	}
	if os == OpenBSD && arch != Amd64 && arch != I386 && arch != Arm && arch != Arm64 {
		return fmt.Errorf("OpenBSD only supports 386, amd64, arm and arm64")
	}
	if os == Plan9 && arch != Amd64 && arch != I386 && arch != Arm {
		return fmt.Errorf("Plan9 only supports 386, amd64, and arm")
	}
	if os == Solaris && arch != Amd64 {
		return fmt.Errorf("Solaris only supports amd64")
	}
	if os == Windows && arch != Amd64 && arch != I386 && arch != Arm && arch != Arm64 {
		return fmt.Errorf("Windows only supports 386, amd64, arm and arm64")
	}
	if os == Linux && arch == Wasm {
		return fmt.Errorf("Wasm is not a valid architecture for Linux")
	}
	return nil
}
