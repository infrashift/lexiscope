// Package version holds the version, commit and build date of lexiscope, and the
// SARIF versions that lexiscope writes.
//
// The build sets Version, Commit and BuildDate with -ldflags -X (see the
// build target of the ZMakefile). A build without these flags, for example
// go install, reads the module version and the VCS stamp of the binary.
package version

import (
	"fmt"
	"runtime/debug"

	"github.com/infrashift/lexiscope/internal/assert"
)

// The build sets these variables with -ldflags -X. An empty value means
// "read the build information of the binary".
var (
	Version   = ""
	Commit    = ""
	BuildDate = ""
)

// DefaultSARIFVersion is the SARIF version of --format sarif.
const DefaultSARIFVersion = "2.1.0"

// SARIFVersions returns the SARIF versions that lexiscope writes.
func SARIFVersions() []string {
	v := []string{"2.1.0", "2.2"}
	assert.True(v[0] == DefaultSARIFVersion, "the default comes first")
	assert.True(len(v) == 2, "lexiscope writes two SARIF versions")
	return v
}

// Info is the version metadata of one binary.
type Info struct {
	Version   string
	Commit    string
	BuildDate string
}

// Get returns the version metadata: the -ldflags values first, then the
// build information of the binary, then "dev" and "unknown".
func Get() Info {
	bi, ok := debug.ReadBuildInfo()
	assert.True(!ok || bi != nil, "ReadBuildInfo gives info when ok")
	info := fromBuildInfo(bi, ok)
	info = overlay(info, Info{Version: Version, Commit: Commit, BuildDate: BuildDate})
	assert.True(info.Version != "" && info.Commit != "" && info.BuildDate != "", "each field has a value")
	return info
}

// fromBuildInfo reads the module version and the VCS stamp.
func fromBuildInfo(bi *debug.BuildInfo, ok bool) Info {
	assert.True(!ok || bi != nil, "the build info exists when ok")
	info := Info{Version: "dev", Commit: "unknown", BuildDate: "unknown"}
	if !ok {
		return info
	}
	if v := bi.Main.Version; v != "" && v != "(devel)" {
		info.Version = v
	}
	dirty := false
	for _, s := range bi.Settings {
		switch s.Key {
		case "vcs.revision":
			info.Commit = s.Value[:min(12, len(s.Value))]
		case "vcs.time":
			info.BuildDate = s.Value
		case "vcs.modified":
			dirty = s.Value == "true"
		}
	}
	if dirty && info.Commit != "unknown" {
		info.Commit += "-dirty"
	}
	assert.True(info.Commit != "", "the commit has a value")
	return info
}

// overlay replaces the fields of base with the non-empty fields of top.
func overlay(base, top Info) Info {
	assert.True(base.Version != "", "the base has a version")
	assert.True(base.Commit != "" && base.BuildDate != "", "the base is complete")
	if top.Version != "" {
		base.Version = top.Version
	}
	if top.Commit != "" {
		base.Commit = top.Commit
	}
	if top.BuildDate != "" {
		base.BuildDate = top.BuildDate
	}
	return base
}

// Details returns "<version> (commit: <commit>, built: <date>)".
func (i Info) Details() string {
	assert.True(i.Version != "", "the version is set")
	assert.True(i.Commit != "" && i.BuildDate != "", "the metadata is set")
	return fmt.Sprintf("%s (commit: %s, built: %s)", i.Version, i.Commit, i.BuildDate)
}

// String returns the version line: "lexiscope version <details>".
func (i Info) String() string {
	s := "lexiscope version " + i.Details()
	assert.True(len(s) > len("lexiscope version "), "the line has details")
	assert.True(s[len(s)-1] == ')', "the line ends with the details")
	return s
}
