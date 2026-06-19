// Package version exposes the build version of mar. The value is overridden at
// build time via -ldflags "-X github.com/statikowsky/mar/internal/version.Version=...".
package version

// Version is the mar build version. It defaults to "dev" for go run / go test
// builds and is set to the git-described version by the Makefile.
var Version = "dev"
