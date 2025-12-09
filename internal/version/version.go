package version

import (
	"fmt"
	"runtime"
)

// These variables can be set at build time using -ldflags
// Example: go build -ldflags "-X tchat/version.Version=1.0.0"
var (
	// Version is the semantic version
	Version = "dev"

	// Commit is the git commit hash
	Commit = "unknown"

	// BuildDate is when the binary was built
	BuildDate = "unknown"
)

// Info contains all version information
type Info struct {
	Version   string
	Commit    string
	BuildDate string
	GoVersion string
	Platform  string
}

// Get returns the current version information
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		BuildDate: BuildDate,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a formatted version string
func (i Info) String() string {
	return fmt.Sprintf(
		"Version:    %s\n"+
			"Commit:     %s\n"+
			"Build Date: %s\n"+
			"Go Version: %s\n"+
			"Platform:   %s",
		i.Version,
		i.Commit,
		i.BuildDate,
		i.GoVersion,
		i.Platform,
	)
}

// Short returns a short version string (just version number)
func (i Info) Short() string {
	return i.Version
}
