package runtime

import (
	_ "embed"
	"encoding/json"
)

type VersionConfig struct {
	NodeVersions []NodeVersion `json:"node"`
	PnpmVersions []string      `json:"pnpm"`
}

type NodeVersion struct {
	Version string `json:"version"`
	LTS     bool   `json:"lts,omitempty"`
}

//go:embed data/versions.json
var versionsJSON []byte

var versions = mustLoadVersions()

func Versions() VersionConfig {
	return versions
}

func mustLoadVersions() VersionConfig {
	var config VersionConfig
	if err := json.Unmarshal(versionsJSON, &config); err != nil {
		panic("parse embedded runtime versions: " + err.Error())
	}
	if len(config.NodeVersions) == 0 || len(config.PnpmVersions) == 0 {
		panic("embedded runtime versions are empty")
	}
	return config
}
