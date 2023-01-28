package apt

import (
	_ "embed"
	"fmt"
	"github.com/codecomet-io/installers/sdk/bash"
	"github.com/codecomet-io/isovaline/isovaline/core/log"
	"github.com/codecomet-io/isovaline/sdk/wrapllb"
	"github.com/codecomet-io/isovaline/sdk/wrapllb/platform"
	"github.com/moby/buildkit/client/llb"
	"strings"
)

/**

Notes:
- lists cache is assumed to be unique per sources.list - indeed, it seems apt-get will nuke it if sources list do not match
- packages cache is more relaxed, amd shared across all builds
- now, we cant segment operations in different layers (eg: update in one layer, download in the next, etc), as some layers
may exist in a previously locally cached version of the image, but the shared volume would not on the buildkit node...

*/

//go:embed sh-art/lock.sh
var shliblock string

//go:embed sh-art/store.sh
var shlibstore string

//go:embed sh-art/apt.sh
var shliblapt string

const (
	sharedStoreMountLocation       = "/_cc/share/apt"
	smallTempFSSize          int64 = 256
	bigTempFSSize            int64 = 1024
)

var (
	sharedStore = &wrapllb.Cache{
		UniqueDescription: fmt.Sprintf("apt-get shared storage, including locks, lists and packages"),
		SharingMode:       wrapllb.ModeShared,
		ReadOnly:          false,
		NoOutput:          false,
	}
)

type AptGet struct {
	// Bash embed
	bash.Bash

	// Config object
	Expert *Config

	// Whether we want the config to be persisted in the State or dropped after use (default)
	PersistConfig bool

	// Whether we want the lists files to be persisted or dropped after use (default)
	PersistLists bool

	// XXX TBD
	logs *wrapllb.Temp
	// cacheuuid string
	SubCom   string
	KeepLogs bool
}

// To be called by Debian source, passing as cacheuuid what is required...
func New(st llb.State) *AptGet {

	ap := &AptGet{
		Bash:   *bash.New(st),
		Expert: NewConfig(),
	}

	// Not technically necessary on every op, but also not worth not mounting it...
	ap.Bash.Cache[sharedStoreMountLocation] = sharedStore

	return ap
}

func (a *AptGet) AddArchitecture(platforms []*platform.Platform) {
	archs := []string{}
	for _, v := range platforms {
		archs = append(archs, string(PlatformToDebianArchitecture(v)))
	}
	name := "dpkg --add-architecture " + strings.Join(archs, " ")
	com := ""
	for _, v := range archs {
		com += fmt.Sprintf("dpkg --add-architecture %s\n", v)
	}

	a.Bash.TMPFSSize = smallTempFSSize
	a.Bash.Run(name, a.wrap(com))

}

// var update_cache_preserver = 0

func (a *AptGet) Update() {
	com := "cc::high::update"
	name := "apt-get update"

	a.Bash.TMPFSSize = smallTempFSSize
	a.Bash.Run(name, a.wrap(com))
}

func (a *AptGet) Do(some []string) {
	com := strings.Join(some, " ")
	name := com

	a.Bash.TMPFSSize = smallTempFSSize
	a.Bash.Run(name, a.wrap(com))
}

func (a *AptGet) Install(packages interface{}) {
	var ipac []string

	switch v := packages.(type) {
	case []*Package:
		for _, vv := range v {
			ipac = append(ipac, vv.toString())
		}
	case *Package:
		ipac = append(ipac, v.toString())
	case []string:
		for _, vv := range v {
			ipac = append(ipac, vv)
		}
	case string:
		ipac = append(ipac, v)
	default:
		log.Fatal().Msgf("invalid packages type %s", v)
	}

	packs := strings.Join(ipac, " ")

	// Download first, keeping the big store usage to a minimum
	com := fmt.Sprintf("cc::high::install %s", packs)
	name := fmt.Sprintf("apt-get install %s", packs)

	a.Bash.TMPFSSize = bigTempFSSize
	a.Bash.Run(name, a.wrap(com))
}

func (a *AptGet) Upgrade() {
	// Download first, keeping the big store usage to a minimum
	com := fmt.Sprintf("cc::high::upgrade")
	name := fmt.Sprintf("apt-get upgrade")

	a.Bash.TMPFSSize = bigTempFSSize
	// a.Bash.Cache[sharedStoreMountLocation] = sharedStore
	a.Bash.Run(name, a.wrap(com))
}

func (a *AptGet) Purge(packages interface{}) {
	var ipac []string

	switch v := packages.(type) {
	case []*Package:
		for _, vv := range v {
			ipac = append(ipac, vv.toString())
		}
	case *Package:
		ipac = append(ipac, v.toString())
	case []string:
		for _, vv := range v {
			ipac = append(ipac, vv)
		}
	case string:
		ipac = append(ipac, v)
	default:
		log.Fatal().Msgf("invalid packages type %s", v)
	}

	packs := strings.Join(ipac, " ")

	// This one does not need any of the mounts
	// com := fmt.Sprintf("cc::apt_get::no_download purge -qq --auto-remove %s", strings.Join(packages, " "))
	com := fmt.Sprintf("cc::high::purge %s", packs)
	name := fmt.Sprintf("apt-get purge %s", packs)

	a.Bash.TMPFSSize = smallTempFSSize
	a.Bash.Run(name, a.wrap(com))
}

func (a *AptGet) wrap(com string) []string {

	return []string{
		shliblock,
		shlibstore,
		shliblapt,
		fmt.Sprintf(`

cc::lock::init %q
cc::apt_get::init "$CC_TMPFS" %q %q %t

cc::apt_get::configure '%s' %t

# Run the command
%s

# Cleanup
## Only there when installing
rm -f /var/cache/ldconfig/aux-cache
rm -f /var/log/dpkg.log
rm -f /var/log/alternatives.log

`,
			sharedStoreMountLocation+"/locks",
			sharedStoreMountLocation+"/lists",
			sharedStoreMountLocation+"/packs",
			a.PersistLists,
			a.Expert.toString(),
			a.PersistConfig,
			com,
		),
	}
}

// fmt.Sprintf("%s", cacheMountLocation),
// XXX hook this up - this should be the last commit timestamp if there is one - otherwise, should be "now"
// "1664825793", // "git -C \"$root\" log -1 --format=\"%at\"",
// XXX this used to be an attempt at making things repeatable
/*

## XXX reproducibility
# All of this may/should disappear with: https://github.com/moby/buildkit/pull/2918
# Looks like somehow we need to fix the timestamp ALSO AFTER we are done

mkdir_timestamp(){
	local pth="$1"
	mkdir -p "$pth"
	epoch="$(date --date "1976-04-14T17:00:00-07:00" +%%s)"

	local remain="$pth"
	while [ "$remain" != "/" ]; do
		touch --no-dereference --date="@$epoch" "$remain"
		remain="$(dirname "$remain")"
	done
}

for mountpoint in %s ; do
	mkdir_timestamp "$mountpoint"
done

# Now, fix timestamp everywhere
epoch="%s"

# Get all entries that have been accessed or modified
entries=()
while read -r file; do
	entries+=("$file")
done < <(
	find / -writable -not -path "/proc*" -not -path "/sys*" -not -path "/dev*" -not -path "/codecomet*" -not -path "/var/log*" \( -newermt "@$epoch" -o -newerat "@$epoch" \)
)

# Reverse sort
indexes=( $(
    for i in "${!entries[@]}" ; do
        printf '%%s %%s %%s\n' $i "${#entries[i]}" "${entries[i]}"
    done | sort -nrk2,2 -rk3 | cut -f1 -d' '
))

sorted=()
for i in "${indexes[@]}" ; do
    sorted+=("${entries[i]}")
done

# Now, fix the timestamp
for entry in "${sorted[@]}"; do
	touch --no-dereference --date="@$epoch" "$entry"
done

# /etc is still a problem: https://github.com/moby/buildkit/issues/3148
# Obviously so are whiteouts...

*/
