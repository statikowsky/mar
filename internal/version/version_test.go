package version

import "testing"

func TestVersionDefaultsToDev(t *testing.T) {
	if Version != "dev" {
		t.Errorf("Version = %q, want dev (test binaries are not built with ldflags)", Version)
	}
}
