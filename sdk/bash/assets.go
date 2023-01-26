package bash

import (
	_ "embed"
	"github.com/codecomet-io/isovaline/sdk/codecomet"
	"github.com/moby/buildkit/client/llb"
	"os"
	"path"
	"strings"
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
)

type assets struct {
	state llb.State

	folder string
}

func (ass *assets) ensureState() error {
	if ass.folder == "" {
		d, e := os.MkdirTemp("", "codecomet.*.shell")
		if e != nil {
			return e
		}
		ass.folder = d
		ass.state = codecomet.From(&codecomet.Local{
			Path: d,
		})
	}
	return nil
}

func (ass *assets) getFile(name string) (string, error) {
	e := ass.ensureState()
	if e != nil {
		return "", e
	}
	if !strings.Contains(name, "*") {
		return path.Join(ass.folder, name), nil
	}
	f, e := os.CreateTemp(ass.folder, name)
	return f.Name(), e
}

func (ass *assets) GetLocalState() (error, llb.State) {
	e := ass.ensureState()
	return e, ass.state
}

func (ass *assets) PackCommander() (error, string) {
	f, e := ass.getFile(commanderFile)
	if e != nil {
		return e, commanderFile
	}
	_, e = os.Stat(f)
	if e != nil {
		e = os.WriteFile(f, []byte(shcodecomet+"\n"+shdebugger+"\n"+shinit), permissions)
	}
	return e, commanderFile
}

func (ass *assets) PackAction(com string) (error, string) {
	f, e := ass.getFile("action.*.sh")
	act := path.Base(f)
	if e != nil {
		return e, act
	}
	e = os.WriteFile(f, []byte(com), permissions)
	if e == nil {
		e = os.Chmod(f, permissions)
	}
	return e, act
}
