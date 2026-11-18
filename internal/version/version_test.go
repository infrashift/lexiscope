package version

import (
	"runtime/debug"
	"slices"
	"testing"
)

func TestFromBuildInfo(t *testing.T) {
	bi := &debug.BuildInfo{
		Main: debug.Module{Version: "v0.3.0"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "0123456789abcdef0123"},
			{Key: "vcs.time", Value: "2026-11-18T06:00:00Z"},
			{Key: "vcs.modified", Value: "true"},
		},
	}
	want := Info{Version: "v0.3.0", Commit: "0123456789ab-dirty", BuildDate: "2026-11-18T06:00:00Z"}
	if got := fromBuildInfo(bi, true); got != want {
		t.Fatalf("fromBuildInfo = %+v, want %+v", got, want)
	}
	devel := &debug.BuildInfo{Main: debug.Module{Version: "(devel)"}}
	if got := fromBuildInfo(devel, true); got != (Info{"dev", "unknown", "unknown"}) {
		t.Fatalf("fromBuildInfo(devel) = %+v", got)
	}
	if got := fromBuildInfo(nil, false); got != (Info{"dev", "unknown", "unknown"}) {
		t.Fatalf("fromBuildInfo(nil) = %+v", got)
	}
}

func TestLdflagsOverlay(t *testing.T) {
	base := Info{"dev", "unknown", "unknown"}
	got := overlay(base, Info{Version: "v1.2.3", Commit: "abc1234"})
	if got != (Info{"v1.2.3", "abc1234", "unknown"}) {
		t.Fatalf("overlay = %+v", got)
	}
}

func TestGetUsesLdflags(t *testing.T) {
	t.Cleanup(func() { Version, Commit, BuildDate = "", "", "" })
	Version, Commit, BuildDate = "v9.9.9", "feedbee", "2026-01-02T03:04:05Z"
	if got, want := Get().String(), "lexis version v9.9.9 (commit: feedbee, built: 2026-01-02T03:04:05Z)"; got != want {
		t.Fatalf("Get().String() = %q, want %q", got, want)
	}
}

func TestSARIFVersions(t *testing.T) {
	if got := SARIFVersions(); !slices.Equal(got, []string{"2.1.0", "2.2"}) {
		t.Fatalf("SARIFVersions() = %v", got)
	}
	SARIFVersions()[0] = "x"
	if SARIFVersions()[0] != DefaultSARIFVersion {
		t.Fatal("a caller can change the list of SARIF versions")
	}
}
