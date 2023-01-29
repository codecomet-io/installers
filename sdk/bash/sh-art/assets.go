package sh_art

import (
	_ "embed"
	"github.com/codecomet-io/installers/sdk"
	"github.com/moby/buildkit/client/llb"
	"strconv"
)

//go:embed debugger.sh
var shdebugger string

//go:embed init.sh
var shinit string

//go:embed codecomet.sh
var shcodecomet string

const (
	permissions   = 0500
	commanderFile = "codecomet"
	library       = "library"
	actionFile    = "action.sh"
)

func Pack(com ...string) (llb.State, []string) {
	states := []llb.State{}
	scripts := []string{}
	scripts = append(scripts, commanderFile)

	ll := []llb.ConstraintsOpt{
		sdk.BackportInternalGroup,
	}

	states = append(states, llb.Scratch().File(llb.Mkfile("/"+commanderFile, permissions, []byte(shcodecomet+"\n"+shdebugger+"\n"+shinit)), ll...))

	for k, v := range com {
		dest := library + strconv.Itoa(k) + ".sh"
		if k == len(com)-1 {
			dest = actionFile
		}
		scripts = append(scripts, dest)
		states = append(states, llb.Scratch().File(llb.Mkfile("/"+dest, permissions, []byte(v)), ll...))
	}

	return llb.Merge(states, ll...), scripts
}
