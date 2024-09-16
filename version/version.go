package version

var (
	version string = "development"
	commit  string
)

func Version() string {
	return version
}

func Commit() string {
	return commit
}

func VersionFull() string {
	result := version
	if len(commit) > 0 {
		result += " (" + commit + ")"
	}
	return result
}
