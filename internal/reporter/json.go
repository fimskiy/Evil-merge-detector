package reporter

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/evilmerge-dev/evil-merge-detector/internal/models"
)

// JSONReporter outputs machine-readable JSON.
type JSONReporter struct {
	Pretty bool
}

func NewJSON(pretty bool) *JSONReporter {
	return &JSONReporter{Pretty: pretty}
}

func (r *JSONReporter) Report(w io.Writer, result *models.ScanResult) error {
	enc := json.NewEncoder(w)
	if r.Pretty {
		enc.SetIndent("", "  ")
	}
	if err := enc.Encode(result); err != nil {
		return fmt.Errorf("encoding JSON: %w", err)
	}
	return nil
}
