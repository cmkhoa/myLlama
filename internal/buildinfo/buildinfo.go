// Package buildinfo identifies development and release binaries.
package buildinfo

import "runtime"

// Set by the release build. Source builds deliberately identify as development.
var Version = "dev"
var Commit = "unknown"
var Date = "unknown"

type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"go_version"`
}

func Current() Info { return Info{Version, Commit, Date, runtime.Version()} }
