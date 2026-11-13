package config

import (
	"encoding/json"
	"flag"
	"io"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/vishenosik/notify/internal/config/tag"
)

const (
	_yaml_ = iota
	_json_
)

type BuildInfo struct {
	ServiceName string      `json:"service_name,omitempty" yaml:"service_name,omitempty"`
	BuildDate   string      `json:"build_date,omitempty" yaml:"build_date,omitempty"`
	GitBranch   string      `json:"git_branch,omitempty" yaml:"git_branch,omitempty"`
	GitCommit   string      `json:"git_commit,omitempty" yaml:"git_commit,omitempty"`
	GoVersion   string      `json:"go_version,omitempty" yaml:"go_version,omitempty"`
	Env         Environment `json:"env,omitempty" yaml:"env,omitempty"`
}

var buildInfo BuildInfo

func init() {
	buildInfo = BuildInfo{
		ServiceName: tag.ServiceName,
		BuildDate:   tag.BuildDate,
		GitBranch:   tag.GitBranch,
		GitCommit:   tag.GitCommit,
		GoVersion:   tag.GoVersion,
		Env:         Environment(tag.Env),
	}
}

func GetBuildInfo() BuildInfo {
	return buildInfo
}

func BuildInfoYaml(writer io.Writer) {
	writeBuildInfo(writer, _yaml_)
}

func BuildInfoJSON(writer io.Writer) {
	writeBuildInfo(writer, _json_)
}

func writeBuildInfo(writer io.Writer, format int) {
	var (
		buf []byte
		err error
	)

	switch format {
	case _yaml_:
		buf, err = yaml.Marshal(buildInfo)
	case _json_:
		buf, err = json.MarshalIndent(buildInfo, "", "  ")
	default:
		return
	}
	if err != nil {
		if _, writeErr := writer.Write([]byte(err.Error())); writeErr != nil {
			return
		}
		return
	}
	if _, err := writer.Write(buf); err != nil {
		return
	}
}

var FlagDebug bool

func Flags() {
	flag.BoolFunc("version", "Show build info", func(s string) error {
		BuildInfoYaml(os.Stdout)
		os.Exit(0)
		return nil
	})
	flag.BoolFunc("level-debug", "Enable debug logging", func(s string) error {
		FlagDebug = true
		return nil
	})
}
