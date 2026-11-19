package cmd

import (
	"runtime/debug"
	"testing"
)

func TestFromBuildInfo(t *testing.T) {
	tests := []struct {
		name         string
		v, c         string
		info         debug.BuildInfo
		wantV, wantC string
	}{
		{
			name: "go install module@version uses the module version",
			v:    "dev", c: "none",
			info:  debug.BuildInfo{Main: debug.Module{Version: "v0.1.1"}},
			wantV: "v0.1.1", wantC: "none",
		},
		{
			name: "local build uses the vcs revision",
			v:    "dev", c: "none",
			info: debug.BuildInfo{Main: debug.Module{Version: "(devel)"},
				Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: "abc1234def5678"}}},
			wantV: "dev", wantC: "abc1234",
		},
		{
			name: "ldflags values win",
			v:    "1.2.3", c: "f00baa1",
			info: debug.BuildInfo{Main: debug.Module{Version: "v9.9.9"},
				Settings: []debug.BuildSetting{{Key: "vcs.revision", Value: "deadbeefcafe"}}},
			wantV: "1.2.3", wantC: "f00baa1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotV, gotC := fromBuildInfo(tt.v, tt.c, &tt.info)
			if gotV != tt.wantV || gotC != tt.wantC {
				t.Errorf("got (%q, %q), want (%q, %q)", gotV, gotC, tt.wantV, tt.wantC)
			}
		})
	}
}
