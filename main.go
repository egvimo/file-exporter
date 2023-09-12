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
}

func newFileCollector(files []string) *fileCollector {
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
	}
}

func (collector *fileCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range collector.fileMetrics {
		ch <- m.metric
	}
}

func (collector *fileCollector) Collect(ch chan<- prometheus.Metric) {
	for _, fileMetric := range collector.fileMetrics {
		var metricValue float64
		if _, err := os.Stat(fileMetric.file); err == nil {
			metricValue = 1
		} else {
			metricValue = 0
		}
		ch <- prometheus.MustNewConstMetric(fileMetric.metric, prometheus.GaugeValue, metricValue)
	}
}

func main() {
	var files arrayFlag
	var addr = flag.String("listen-address", ":8080", "The address to listen on for HTTP requests.")
	flag.Var(&files, "file", "File to export metric for.")

	flag.Parse()

	if len(files) == 0 {
		log.Fatal("No files provided")
		return
	}

	log.Printf("Initializing exporter for files: %v", files)

	fileCollector := newFileCollector(files)

	reg := prometheus.NewRegistry()
	reg.MustRegister(fileCollector)

	http.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{EnableOpenMetrics: false}))

	log.Printf("Serving on %s", *addr)
	log.Fatal(http.ListenAndServe(*addr, nil))
}
