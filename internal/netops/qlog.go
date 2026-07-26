package netops

import (
	"encoding/json"
	"io"
	"os"
)

// QLogEvent represents a single event in a qlog file.
type QLogEvent struct {
	RelativeTime float64         `json:"time"`
	Name         string          `json:"name"`
	Data         json.RawMessage `json:"data"`
}

// QLogTrace represents a trace in a qlog file.
type QLogTrace struct {
	VantagePoint map[string]string `json:"vantage_point"`
	Title        string            `json:"title"`
	Events       []QLogEvent       `json:"events"`
}

// QLogFile represents the top-level qlog structure.
type QLogFile struct {
	QLogVersion string      `json:"qlog_version"`
	Traces      []QLogTrace `json:"traces"`
}

// ParseQLog reads a .qlog file and returns the structured data.
func ParseQLog(path string) (*QLogFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var qlog QLogFile
	if err := json.Unmarshal(data, &qlog); err != nil {
		return nil, err
	}

	return &qlog, nil
}
