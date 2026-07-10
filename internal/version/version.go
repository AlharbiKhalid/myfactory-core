// Package version holds build metadata injected at release time via
// -ldflags "-X github.com/AlharbiKhalid/myfactory-core/internal/version.Version=v1.2.3 ...".
package version

var (
	// Version is the MyFactory release version. "dev" for local builds.
	Version = "dev"
	// Commit is the git commit the binary was built from.
	Commit = "unknown"
	// Date is the build date (UTC, RFC 3339).
	Date = "unknown"
)
