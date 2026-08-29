package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

const (
	nodeIndexURL    = "https://nodejs.org/dist/index.json"
	nodeScheduleURL = "https://raw.githubusercontent.com/nodejs/Release/main/schedule.json"
	pnpmRegistryURL = "https://registry.npmjs.org/pnpm"
	dataPath        = "internal/runtime/data/versions.json"
)

type nodeRelease struct {
	Version string `json:"version"`
}
type nodeSchedule struct {
	LTS json.RawMessage `json:"lts"`
}
type pnpmMetadata struct {
	Versions map[string]json.RawMessage `json:"versions"`
}
type nodeVersion struct {
	Version string `json:"version"`
	LTS     bool   `json:"lts,omitempty"`
}
type runtimeConfig struct {
	Node []nodeVersion `json:"node"`
	Pnpm []string      `json:"pnpm"`
}

var client = http.Client{Timeout: 30 * time.Second}

func main() {
	var releases []nodeRelease
	var schedule map[string]nodeSchedule
	var metadata pnpmMetadata
	for _, request := range []struct {
		url  string
		out  any
		name string
	}{
		{
			nodeIndexURL,
			&releases,
			"Node.js releases",
		},
		{
			nodeScheduleURL,
			&schedule,
			"Node.js schedule",
		},
		{
			pnpmRegistryURL,
			&metadata,
			"pnpm releases",
		},
	} {
		if err := fetch(request.url, request.out); err != nil {
			fail("fetch "+request.name, err)
		}
	}

	nodeValues := make([]string, 0, len(releases))
	for _, release := range releases {
		nodeValues = append(nodeValues, release.Version)
	}
	nodeReleases := latest(nodeValues, 18)
	node := make([]nodeVersion, len(nodeReleases))
	for index, version := range nodeReleases {
		node[index] = nodeVersion{
			Version: version.String(),
			LTS:     isLTS(schedule[fmt.Sprintf("v%d", version.Major())].LTS),
		}
	}

	pnpmValues := make([]string, 0, len(metadata.Versions))
	for version := range metadata.Versions {
		pnpmValues = append(pnpmValues, version)
	}
	pnpm := latest(pnpmValues, 10)
	if len(node) == 0 || len(pnpm) == 0 {
		fail("build runtime config", fmt.Errorf("upstream metadata contains no supported versions"))
	}

	config := runtimeConfig{Node: node, Pnpm: make([]string, len(pnpm))}
	for index, version := range pnpm {
		config.Pnpm[index] = version.String()
	}
	data, err := json.MarshalIndent(config, "", strings.Repeat(" ", 2))
	if err != nil {
		fail("encode runtime config", err)
	}
	if err := os.WriteFile(dataPath, append(data, '\n'), 0o644); err != nil {
		fail("write runtime config", err)
	}
	fmt.Printf("Updated %s\n", dataPath)
}

func fetch(url string, target any) error {
	response, err := client.Get(url)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("server returned %s", response.Status)
	}
	return json.NewDecoder(response.Body).Decode(target)
}

func latest(values []string, minimumMajor uint64) []*semver.Version {
	versions := make(map[uint64]*semver.Version)
	for _, value := range values {
		version, err := stable(value)
		if err != nil || version.Major() < minimumMajor {
			continue
		}
		if current := versions[version.Major()]; current == nil || version.GreaterThan(current) {
			versions[version.Major()] = version
		}
	}
	result := make([]*semver.Version, 0, len(versions))
	for _, version := range versions {
		result = append(result, version)
	}
	sort.Slice(result, func(left, right int) bool {
		return result[left].GreaterThan(result[right])
	})
	return result
}

func stable(value string) (*semver.Version, error) {
	version, err := semver.StrictNewVersion(strings.TrimPrefix(strings.TrimSpace(value), "v"))
	if err != nil || version.Prerelease() != "" || version.Metadata() != "" {
		return nil, fmt.Errorf("not a stable semantic version")
	}
	return version, nil
}

func isLTS(value json.RawMessage) bool {
	return strings.HasPrefix(strings.TrimSpace(string(value)), `"`)
}

func fail(action string, err error) {
	fmt.Fprintf(os.Stderr, "update runtime: %s: %v\n", action, err)
	os.Exit(1)
}
