package logs

import (
	// builtin
	"encoding/json"
	"log/slog"

	"gopkg.in/yaml.v3"
)

const (
	AttrError = "error"
)

func Error(err error) slog.Attr {
	if err == nil {
		return slog.String(AttrError, "error is nil")
	}

	_, jsonOk := err.(json.Marshaler)
	_, yamlOk := err.(yaml.Marshaler)

	if jsonOk && yamlOk {
		return slog.Any(AttrError, err)
	}

	return slog.String(AttrError, err.Error())
}
