package bash

import (
	_ "embed"
	sh_art "github.com/codecomet-io/installers/sdk/bash/sh-art"
	"github.com/codecomet-io/isovaline/isovaline/core/log"
	"github.com/codecomet-io/isovaline/sdk/codecomet"
	"github.com/codecomet-io/isovaline/sdk/wrapllb"
	"github.com/moby/buildkit/client/llb"
	"net"
	"os"
)

const tempMountLocation = "/root"
const scriptsMountLocation = "/_cc/bash"
const defaultTempFSSize = 128
const defaultHostname = "codecomet-bash"

const (
	MUTE int = iota
	ERROR
	WARNING
	INFO
	DEBUG
)

var (
	env_debug       = os.Getenv("CODECOMET_DEBUG")
	env_debug_grace = os.Getenv("CODECOMET_DEBUGGER_WAIT")
	env_debug_core  = os.Getenv("CODECOMET_DEBUG_CORE")
	env_debug_level = os.Getenv("CODECOMET_LOG_LEVEL") // strconv.Itoa
	env_no_color    = os.Getenv("NO_COLOR")
	env_debug_live  = "true"
)

func init() {
	v, b := os.LookupEnv("CODECOMET_DEBUGGER_LIVE")
	if b {
		env_debug_live = v
	}
}

// Get preferred outbound ip of this machine
func getOutboundIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		log.Fatal().Msgf("%s", err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

var (
	BaseEnv = map[string]string{
		// Basic stuff
		"DEBIAN_FRONTEND": "noninteractive",
		"TERM":            "xterm",
		"LANG":            "C.UTF-8",
		"LC_ALL":          "C.UTF-8",
		// XXX timezone does not work rn
		"TZ": "America/Los_Angeles",

		// Plumbing
		"NO_COLOR": env_no_color,
		"TMPDIR":   tempMountLocation + "/tmp",

		// Internal cc stuff
		"CC_TMPFS":          tempMountLocation,
		"CC_DEBUGGER_GRACE": env_debug_grace,
		"CC_DEBUGGER_PORT":  "6666",
		"CC_STDOUT":         "stdout",
		"CC_STDERR":         "stderr",

		// Internal debugging only
		"CC_DEBUG_CORE":  env_debug_core,
		"CC_DEBUG_LEVEL": env_debug_level, // strconv.Itoa(INFO),
		"CC_DEBUG_LIVE":  env_debug_live,
	}
)

type Bash struct {
	// State we are manipulating
	State llb.State

	// Whether the state is readonly
	ReadOnly bool
	// Toggle debug on
	Debug bool
	// Toggle strict behavior on
	Strict bool
	// Temp size
	TMPFSSize int64

	// Environment
	Env map[string]string

	// Hostname
	Hostname string

	// Temp mounts
	Temp map[wrapllb.Target]*wrapllb.Temp
	// Cache mounts
	Cache map[wrapllb.Target]*wrapllb.Cache
	// State mounts
	Mount map[wrapllb.Target]*wrapllb.State

	// Which dir to start in
	Dir string
	// Group
	Group *wrapllb.Group
	// Advanced bash options
	Expert *Config
}

func New(src llb.State) *Bash {
	// Copy over base env
	env := make(map[string]string)
	for k, v := range BaseEnv {
		env[k] = v
	}

	src = src.Hostname(defaultHostname)

	cnf := &Config{
		// Default bash behavior
		BraceExpand:         true,
		Emacs:               true,
		Hashall:             true,
		InteractiveComments: true,

		// Job control for live debuggers
		Monitor: true,
		// XXX does it allow the debugger to access the history?
		History: true,
	}

	// Default behavior is strict, no debug, in the tmpfs dir
	return &Bash{
		Strict: true,
		Debug:  env_debug == "true",
		Expert: cnf,
		Env:    env,
		State:  src,
		Dir:    tempMountLocation,
		Mount:  make(map[wrapllb.Target]*wrapllb.State),
		Temp:   make(map[wrapllb.Target]*wrapllb.Temp),
		Cache:  make(map[wrapllb.Target]*wrapllb.Cache),
	}
}

func (bsh *Bash) pack(libs []string, action string) (llb.State, []string) {
	// Toggle on debug and strict
	if bsh.Debug {
		bsh.Expert.XTrace = true
	}
	// Note: if one wants only some of these, Strict should be set to false
	if bsh.Strict {
		bsh.Expert.ErrExit = true
		bsh.Expert.ErrTrace = true
		bsh.Expert.FuncTrace = true
		bsh.Expert.NoUnset = true
		bsh.Expert.PipeFail = true
	}

	return sh_art.Pack(append(libs, "#!/usr/bin/env bash\n"+bsh.Expert.toString()+"\n"+action))
}

func (bsh *Bash) Run(name string, com []string /*, pg llb.RunOption*/) {
	// No name means the user script is the name
	act := com[len(com)-1]
	com = com[:len(com)-1]
	if name == "" {
		name = act
	}

	if bsh.Env["CC_DEBUGGER_IP"] == "" {
		bsh.Env["CC_DEBUGGER_IP"] = getOutboundIP().String()
	}

	if bsh.Hostname != "" {
		bsh.State = bsh.State.Hostname(bsh.Hostname)
	}

	sz := bsh.TMPFSSize
	if sz == 0 {
		sz = defaultTempFSSize
	}
	bsh.Temp[tempMountLocation] = &wrapllb.Temp{
		Size: sz * 1024 * 1024,
	}

	sst, scripts := bsh.pack(com, act)
	bsh.Mount[scriptsMountLocation] = &wrapllb.State{
		ReadOnly: true,
		NoOutput: true,
		Source:   sst,
		Path:     "/",
	}
	for k, v := range scripts {
		scripts[k] = scriptsMountLocation + "/" + v
	}

	ce := &codecomet.Exec{
		// Stuff in the base
		Base: codecomet.Base{
			CustomName: name,
			Group:      bsh.Group,
		},

		// And the rest
		Dir:      bsh.Dir,
		Args:     scripts,
		Temp:     bsh.Temp,
		Cache:    bsh.Cache,
		Mount:    bsh.Mount,
		Envs:     bsh.Env,
		ReadOnly: bsh.ReadOnly,
	}
	bsh.State = codecomet.Execute(bsh.State, ce)
	// Carry over mounts?
	// bsh.Mount = ce.Mount
}
