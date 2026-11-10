package mapper

import (
	"github.com/vishenosik/notify/internal/app/common"
	"github.com/vishenosik/notify/internal/config"
)

const (
	statusOK    = "ok"
	statusError = "error"
)

type HealthCheckResponse struct {
	Status string `json:"status"`
}

func ToHealthCheckResponse(response *common.HealthCheckResponse) *HealthCheckResponse {
	return &HealthCheckResponse{
		Status: toStatus(response.Healthy),
	}
}

type HealthCheckResponseVerbose struct {
	Status     string            `json:"status"`
	Subsystems map[string]string `json:"subsystems"`
}

func ToHealthCheckResponseVerbose(response *common.HealthCheckResponse) *HealthCheckResponseVerbose {
	subsystems := make(map[string]string)

	for name, healthy := range response.Subsystems {
		if name == "" {
			continue
		}
		subsystems[name] = toStatus(healthy)
	}

	return &HealthCheckResponseVerbose{
		Status:     toStatus(response.Healthy),
		Subsystems: subsystems,
	}
}

func toStatus(healthy bool) string {
	if healthy {
		return statusOK
	}
	return statusError
}

type BuildInfoResponse struct {
	ServiceName string `json:"service_name"`
	BuildDate   string `json:"build_date"`
	GitBranch   string `json:"git_branch,omitempty"`
	GitCommit   string `json:"git_commit,omitempty"`
	GoVersion   string `json:"go_version,omitempty"`
	Env         string `json:"env,omitempty"`
}

func NewBuildInfoResponse(buildInfo config.BuildInfo) *BuildInfoResponse {
	return &BuildInfoResponse{
		ServiceName: buildInfo.ServiceName,
		BuildDate:   buildInfo.BuildDate,
		GitBranch:   buildInfo.GitBranch,
		GitCommit:   buildInfo.GitCommit,
		GoVersion:   buildInfo.GoVersion,
		Env:         buildInfo.Env.String(),
	}
}
