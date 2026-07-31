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

// Test tool dependencies are imported here to prevent `go mod tidy` from
// removing them. They are used in test files throughout the project.
//
//   - github.com/stretchr/testify: assertion library for all Go tests
//   - github.com/golang/mock/mockgen: optional mock code generator

import (
	_ "github.com/prometheus/client_golang/prometheus"
	_ "github.com/prometheus/client_golang/prometheus/promhttp"
	_ "github.com/stretchr/testify/assert"
	_ "github.com/stretchr/testify/require"
)
