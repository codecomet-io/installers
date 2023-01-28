package debian

type Version string

const (
	Jessie   Version = "jessie-slim"
	Stretch  Version = "stretch-slim"
	Buster   Version = "buster-slim"
	Bullseye Version = "bullseye-slim"
	Bookworm Version = "bookworm-slim"
	Sid      Version = "sid-slim"
)
