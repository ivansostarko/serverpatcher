package version

import "fmt"

var (
	// Set via -ldflags at build time.
	Version   = "0.1.0"
	Commit    = "dev"
	BuildDate = "unknown"
)

func VersionString() string {
	return fmt.Sprintf("%s (commit=%s build=%s)", Version, Commit, BuildDate)
}
