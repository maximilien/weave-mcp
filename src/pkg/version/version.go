// SPDX-License-Identifier: MIT
// Copyright (c) 2025 dr.max

package version

import (
	"fmt"
	"runtime"
)

// Version information
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

// Info holds version information
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit"`
	BuildTime string `json:"build_time"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// Get returns version information
func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildTime: BuildTime,
		GoVersion: runtime.Version(),
		Platform:  fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}

// String returns a formatted version string
func String() string {
	info := Get()
	return fmt.Sprintf("Weave MCP Server\nVersion: %s\nGit Commit: %s\nBuild Time: %s\nGo Version: %s\nPlatform: %s",
		info.Version, info.GitCommit, info.BuildTime, info.GoVersion, info.Platform)
}
