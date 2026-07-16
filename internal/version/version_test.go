package version

import (
	"runtime/debug"
	"testing"
)

func TestVersionDefaultsToDev(t *testing.T) {
	if Version != "dev" {
		t.Errorf("Version = %q, want dev (test binaries are not built with ldflags)", Version)
	}
}

func TestResolveBuildVersion(t *testing.T) {
	tests := []struct {
		name, linked, embedded, want string
	}{
		{"linked version wins", "v1.2.3", "v9.9.9", "v1.2.3"},
		{"go install version", "dev", "v1.2.3", "v1.2.3"},
		{"local build", "dev", "(devel)", "dev"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := &debug.BuildInfo{Main: debug.Module{Version: tt.embedded}}
			if got := resolveBuildVersion(tt.linked, info); got != tt.want {
				t.Errorf("resolveBuildVersion(%q, %q) = %q, want %q", tt.linked, tt.embedded, got, tt.want)
			}
		})
	}
}
