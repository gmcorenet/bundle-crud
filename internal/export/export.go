package export

import (
	"encoding/csv"
	"fmt"
	"strings"
)

type ExportRequest struct {
	Resource string
	Action   string
	Content  string
}

type ExportResult struct {
	Filename    string
	ContentType string
	Data        []byte
}

type Exporter interface {
	Format() string
	Export(req ExportRequest) (*ExportResult, error)
}

type ExportRegistry struct {
	exporters map[string]Exporter
}

func NewExportRegistry() *ExportRegistry {
	r := &ExportRegistry{
		exporters: make(map[string]Exporter),
	}
	r.exporters["txt"] = &TextExporter{}
	r.exporters["csv"] = &CSVExporter{}
	return r
}

func (r *ExportRegistry) Register(e Exporter) {
	r.exporters[e.Format()] = e
}

func (r *ExportRegistry) Get(format string) (Exporter, bool) {
	e, ok := r.exporters[format]
	return e, ok
}

func (r *ExportRegistry) Formats() []string {
	formats := make([]string, 0, len(r.exporters))
	for f := range r.exporters {
		formats = append(formats, f)
	}
	return formats
}

var globalRegistry *ExportRegistry

func GlobalRegistry() *ExportRegistry {
	if globalRegistry == nil {
		globalRegistry = NewExportRegistry()
	}
	return globalRegistry
}

type TextExporter struct{}

func (t *TextExporter) Format() string { return "txt" }

func (t *TextExporter) Export(req ExportRequest) (*ExportResult, error) {
	return &ExportResult{
		Filename:    req.Resource + ".txt",
		ContentType: "text/plain; charset=utf-8",
		Data:        []byte(req.Content),
	}, nil
}

type CSVExporter struct{}

func (c *CSVExporter) Format() string { return "csv" }

func (c *CSVExporter) Export(req ExportRequest) (*ExportResult, error) {
	var buf strings.Builder
	w := csv.NewWriter(&buf)

	lines := strings.Split(req.Content, "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if err := w.Write(fields); err != nil {
			return nil, fmt.Errorf("csv write error: %w", err)
		}
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return nil, fmt.Errorf("csv flush error: %w", err)
	}

	return &ExportResult{
		Filename:    req.Resource + ".csv",
		ContentType: "text/csv; charset=utf-8",
		Data:        []byte(buf.String()),
	}, nil
}
