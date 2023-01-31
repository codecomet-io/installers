package debian

type Version string

const (
	Jessie    Version = "jessie"
	Stretch   Version = "stretch"
	Buster    Version = "buster"
	Bullseye  Version = "bullseye"
	Bookworm  Version = "bookworm"
	Sid       Version = "sid"
	OldStable Version = "oldstable"
	Stable    Version = "stable"
	Testing   Version = "testing"
	Unstable  Version = "unstable"
)
