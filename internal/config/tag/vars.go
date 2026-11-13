package tag

const unset = "not configured"

var (
	ServiceName string = unset
	BuildDate   string = unset
	GitBranch   string = unset
	GitCommit   string = unset
	GoVersion   string = unset
	Env         string = "production"
)
