package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type FileSystemChecker interface {
	FileExists(string) bool
}

type OSFileSystem struct{}

func (OSFileSystem) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

type arrayFlag []string

func (i *arrayFlag) String() string {
	return fmt.Sprintf("%v", *i)
}

func (i *arrayFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

type fileMetric struct {
	file   string
	metric *prometheus.Desc
}

type fileCollector struct {
	fileMetrics []fileMetric
	fsChecker   FileSystemChecker
}

func NewFileCollector(files []string, fsChecker FileSystemChecker) *fileCollector {
	fileMetrics := make([]fileMetric, 0, len(files))

	for _, file := range files {
		metric := prometheus.NewDesc(
			"file_exists",
			"File or directory exists",
			nil,
			map[string]string{"file": file},
		)
		fileMetrics = append(fileMetrics, fileMetric{file, metric})
	}

	return &fileCollector{
		fileMetrics: fileMetrics,
		fsChecker:   fsChecker,
	}
}

func (collector *fileCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range collector.fileMetrics {
		ch <- m.metric
	}
}

func (collector *fileCollector) Collect(ch chan<- prometheus.Metric) {
	for _, fileMetric := range collector.fileMetrics {
		metricValue := 0.0
		if collector.fsChecker.FileExists(fileMetric.file) {
			metricValue = 1.0
		}
		ch <- prometheus.MustNewConstMetric(fileMetric.metric, prometheus.GaugeValue, metricValue)
	}
}

func StartMetricsServer(addr string, files []string) error {
	if len(files) == 0 {
		return fmt.Errorf("No files provided")
	}

	log.Printf("Initializing exporter for files: %v", files)

	fileCollector := NewFileCollector(files, OSFileSystem{})

	reg := prometheus.NewRegistry()
	reg.MustRegister(fileCollector)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	log.Printf("Serving on %s", addr)
	return http.ListenAndServe(addr, nil)
}

func main() {
	var files arrayFlag
	var addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	flag.Var(&files, "file", "File to export metric for.")

	flag.Parse()

	if err := StartMetricsServer(*addr, files); err != nil {
		log.Fatal(err)
	}
}
