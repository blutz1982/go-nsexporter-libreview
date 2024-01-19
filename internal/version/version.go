package version

import (
	"fmt"
	"runtime"
)

var (
	version string = "development"
)

const CmdVersionTemplate = `{{with .Name}}{{printf "%s " .}}{{end}}{{printf "%s" .Version}}
`

// BuildInfo describes the compile time information.
type BuildInfo struct {
	// Version is the current semver.
	AppVersion string `json:"app_version,omitempty"`
	// GoVersion is the version of the Go compiler used.
	GoVersion string `json:"go_version,omitempty"`
}

// Get returns build info
func Get() BuildInfo {
	v := BuildInfo{
		AppVersion: version,
		GoVersion:  runtime.Version(),
	}

	return v
}

func FormatVersion() string {
	return fmt.Sprintf("%#v", Get())
}
