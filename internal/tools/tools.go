//go:build tools

package tools

// This file prevents `go mod tidy` from removing prometheus/client_golang
// (and its transitive dependencies) from go.mod when the "prometheus" build
// tag is not active. The actual import is behind //go:build prometheus in
// internal/common/metrics_exporter.go.
//
// To build with Prometheus support: go build -tags prometheus
//
// References:
//   - https://go.dev/wiki/Modules#how-can-i-track-tool-dependencies-for-a-module

import (
	_ "github.com/prometheus/client_golang/prometheus"
	_ "github.com/prometheus/client_golang/prometheus/promhttp"
)
