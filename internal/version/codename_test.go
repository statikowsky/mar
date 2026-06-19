package version

import (
	"strings"
	"testing"
)

func TestCodenameDeterministic(t *testing.T) {
	if Codename("v0.3.0") != Codename("v0.3.0") {
		t.Fatal("Codename not deterministic for the same version")
	}
}

func TestCodenameIsTwoWordsFromBanks(t *testing.T) {
	adj := map[string]bool{}
	for _, a := range codenameAdjectives {
		adj[a] = true
	}
	noun := map[string]bool{}
	for _, n := range codenameNouns {
		noun[n] = true
	}
	for _, v := range []string{"dev", "v0.1.0", "v1.2.3", "ecc9ac8-dirty", ""} {
		name := Codename(v)
		parts := strings.SplitN(name, " ", 2)
		if len(parts) != 2 || !adj[parts[0]] || !noun[parts[1]] {
			t.Errorf("Codename(%q) = %q, want <adjective> <noun> drawn from the banks", v, name)
		}
	}
}

func TestDisplayIncludesVersionAndCodename(t *testing.T) {
	got := Display()
	if !strings.HasPrefix(got, Version+" \"") || !strings.Contains(got, Codename(Version)) {
		t.Errorf("Display() = %q, want version followed by the quoted codename", got)
	}
}
