package apt

import (
	"fmt"
	"github.com/codecomet-io/isovaline/sdk/wrapllb/platform"
)

// Package represent a Debian deb
type Package struct {
	Name     string
	Version  string
	Platform *platform.Platform
}

// toString serializes the package in a way usable by apt-get command-line
func (pkg *Package) toString() string {
	pack := pkg.Name
	if pkg.Platform != nil {
		pack = fmt.Sprintf("%s:%s", pack, PlatformToDebianArchitecture(pkg.Platform))
	}
	if pkg.Version != "" {
		pack = fmt.Sprintf("%s=%s", pack, pkg.Version)
	}
	return pack
}
