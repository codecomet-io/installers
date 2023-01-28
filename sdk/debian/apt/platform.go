package apt

import (
	"github.com/codecomet-io/isovaline/isovaline/core/log"
	"github.com/codecomet-io/isovaline/sdk/wrapllb/platform"
)

type Architecture string

const (
	architectureArm64 Architecture = "arm64"
	architectureAmd64              = "amd64"
	architectureArmhf              = "armhf"
	architectureArmel              = "armel"
)

func PlatformToDebianArchitecture(plt *platform.Platform) Architecture {
	if plt.OS != platform.Linux {
		log.Fatal().Msgf("only linux is supported and not %s", plt.OS)
	}
	switch plt.Architecture {
	case platform.Arm64:
		return architectureArm64
	case platform.Amd64:
		return architectureAmd64
	case platform.Arm:
		switch plt.Variant {
		case platform.V7:
			return architectureArmhf
		case platform.V6:
			return architectureArmel
		}
	default:
		log.Fatal().Msgf("unsupported achitecture %s %s %s", plt.OS, plt.Architecture, plt.Variant)
	}
	return ""
}
