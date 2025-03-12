package main

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
)

type MockFileSystem struct {
	existingFiles map[string]bool
}

func (m MockFileSystem) FileExists(path string) bool {
	return m.existingFiles[path]
}

func TestFileCollector(t *testing.T) {
	mockFS := MockFileSystem{existingFiles: map[string]bool{
		"/path/to/file1": true,
		"/path/to/file2": false,
	}}

	files := []string{"/path/to/file1", "/path/to/file2"}
	collector := NewFileCollector(files, mockFS)

	ch := make(chan prometheus.Metric, len(files))
	collector.Collect(ch)

	metrics := map[string]float64{}

	for i := 0; i < len(files); i++ {
		metric := <-ch

		dtoMetric := &dto.Metric{}
		if err := metric.Write(dtoMetric); err != nil {
			t.Fatalf("Failed to read metric: %v", err)
		}

		var fileLabel string
		for _, label := range dtoMetric.GetLabel() {
			if label.GetName() == "file" {
				fileLabel = label.GetValue()
				break
			}
		}

		metrics[fileLabel] = dtoMetric.GetGauge().GetValue()
	}

	assert.Equal(t, map[string]float64{
		"/path/to/file1": 1.0,
		"/path/to/file2": 0.0,
	}, metrics)
}
