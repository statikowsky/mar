package homebrew

import (
	"strings"
	"testing"
)

func TestRenderFormula(t *testing.T) {
	checksums := strings.NewReader(strings.Join([]string{
		"aaa111  mar_darwin_arm64.tar.gz",
		"bbb222  mar_darwin_amd64.tar.gz",
		"ccc333  mar_linux_arm64.tar.gz",
		"ddd444  mar_linux_amd64.tar.gz",
		"eee555  mar_windows_amd64.zip",
	}, "\n"))

	got, err := RenderFormula("v0.2.0", checksums)
	if err != nil {
		t.Fatalf("RenderFormula: %v", err)
	}
	for _, want := range []string{
		`version "0.2.0"`,
		`url "https://github.com/statikowsky/mar/releases/download/v0.2.0/mar_darwin_arm64.tar.gz"`,
		`sha256 "aaa111"`,
		`url "https://github.com/statikowsky/mar/releases/download/v0.2.0/mar_darwin_amd64.tar.gz"`,
		`sha256 "bbb222"`,
		`url "https://github.com/statikowsky/mar/releases/download/v0.2.0/mar_linux_arm64.tar.gz"`,
		`sha256 "ccc333"`,
		`url "https://github.com/statikowsky/mar/releases/download/v0.2.0/mar_linux_amd64.tar.gz"`,
		`sha256 "ddd444"`,
		`bin.install "mar"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("formula missing %q:\n%s", want, got)
		}
	}
}

func TestRenderFormulaRequiresAllReleaseArchives(t *testing.T) {
	_, err := RenderFormula("v0.2.0", strings.NewReader("aaa111  mar_darwin_arm64.tar.gz\n"))
	if err == nil || !strings.Contains(err.Error(), "mar_darwin_amd64.tar.gz") {
		t.Fatalf("RenderFormula error = %v, want missing archive", err)
	}
}

func TestRenderFormulaRejectsBlankTag(t *testing.T) {
	_, err := RenderFormula(" ", strings.NewReader(""))
	if err == nil || !strings.Contains(err.Error(), "tag is required") {
		t.Fatalf("RenderFormula error = %v, want tag required", err)
	}
}
