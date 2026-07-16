// Package version exposes the build version of mar. The value is overridden at
// build time via -ldflags "-X github.com/statikowsky/mar/internal/version.Version=...".
package version

import "runtime/debug"

// Version is the mar build version. It defaults to "dev" for go run / go test
// builds and is set to the git-described version by the Makefile.
var Version = "dev"

func init() {
	info, _ := debug.ReadBuildInfo()
	Version = resolveBuildVersion(Version, info)
}

func resolveBuildVersion(linked string, info *debug.BuildInfo) string {
	if linked != "dev" || info == nil || info.Main.Version == "" || info.Main.Version == "(devel)" {
		return linked
	}
	return info.Main.Version
}
