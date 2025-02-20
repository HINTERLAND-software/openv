package version

// Version information
var (
	// Version is the current version of the application.
	// This should be set using -ldflags at build time
	Version = "dev"

	// CommitHash is the git commit hash at build time
	CommitHash = "unknown"

	// BuildTime is the build timestamp
	BuildTime = "unknown"
)

// Info returns version, commit hash and build time
func Info() string {
	return Version + " (" + CommitHash + ") built at " + BuildTime
}
