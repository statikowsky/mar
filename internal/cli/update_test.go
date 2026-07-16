package cli

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/statikowsky/mar/internal/version"
)

func TestUpdateCommandReportsAvailableRelease(t *testing.T) {
	stubLatestRelease(t, http.StatusOK, `{"tag_name":"v0.6.0","html_url":"https://github.com/statikowsky/mar/releases/tag/v0.6.0"}`)
	setCurrentVersion(t, "v0.5.2")

	out, err := runCmd(t, "", "update")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"v0.5.2 → v0.6.0", "brew upgrade mar", "go install github.com/statikowsky/mar@latest"} {
		if !strings.Contains(out, want) {
			t.Errorf("update output missing %q:\n%s", want, out)
		}
	}
}

func TestUpdateCommandJSON(t *testing.T) {
	stubLatestRelease(t, http.StatusOK, `{"tag_name":"v0.5.2","html_url":"https://github.com/statikowsky/mar/releases/tag/v0.5.2"}`)
	setCurrentVersion(t, "v0.5.2")

	out, err := runCmd(t, "", "update", "--json")
	if err != nil {
		t.Fatal(err)
	}
	got := decodeJSON(t, out)
	if got["current"] != "v0.5.2" || got["latest"] != "v0.5.2" || got["update_available"] != false {
		t.Errorf("got %v", got)
	}
}

func TestUpdateCommandRejectsFailedReleaseRequest(t *testing.T) {
	stubLatestRelease(t, http.StatusForbidden, `{}`)
	if _, err := runCmd(t, "", "update"); err == nil || !strings.Contains(err.Error(), "403 Forbidden") {
		t.Fatalf("error = %v, want GitHub status", err)
	}
}

func stubLatestRelease(t *testing.T, status int, body string) {
	t.Helper()
	old := updateClient
	updateClient = &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: status,
			Status:     fmt.Sprintf("%d %s", status, http.StatusText(status)),
			Body:       io.NopCloser(strings.NewReader(body)),
		}, nil
	})}
	t.Cleanup(func() {
		updateClient = old
	})
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func setCurrentVersion(t *testing.T, current string) {
	t.Helper()
	old := version.Version
	version.Version = current
	t.Cleanup(func() { version.Version = old })
}
