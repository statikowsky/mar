package version

import "hash/fnv"

// codenameAdjectives and codenameNouns are the word banks for release
// codenames. A name is one adjective plus one nautical noun ("Calm Harbor"),
// fitting mar's name (mar = sea). The pair is derived deterministically from
// the version string, so every build — tagged or dev — gets a stable,
// memorable name. 24 x 26 = 624 combinations.
var codenameAdjectives = []string{
	"Amber", "Bright", "Calm", "Cold", "Deep", "Distant", "Foggy", "Grey",
	"Hidden", "Lonely", "North", "Open", "Quiet", "Restless", "Rolling", "Salt",
	"Shallow", "Silver", "Still", "Stormy", "Tidal", "Wild", "Windward", "South",
}

var codenameNouns = []string{
	"Anchor", "Beacon", "Channel", "Cove", "Current", "Drift", "Fathom", "Gale",
	"Gulf", "Harbor", "Haven", "Keel", "Lagoon", "Mistral", "Narrows", "Passage",
	"Quay", "Reef", "Shoal", "Sound", "Strait", "Swell", "Tide", "Trench",
	"Wake", "Wave",
}

// Codename returns the memorable two-word name for a version string. The same
// version always maps to the same name; different versions are spread across
// the combination space by hashing.
func Codename(v string) string {
	h := fnv.New32a()
	h.Write([]byte(v))
	sum := h.Sum32()
	adj := codenameAdjectives[sum%uint32(len(codenameAdjectives))]
	noun := codenameNouns[(sum/uint32(len(codenameAdjectives)))%uint32(len(codenameNouns))]
	return adj + " " + noun
}

// Display combines the build version with its codename for human-facing
// output, e.g. `v0.3.0 "Rolling Reef"`.
func Display() string {
	return Version + " \"" + Codename(Version) + "\""
}
