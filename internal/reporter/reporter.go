package reporter

import (
	"io"

	"github.com/fimskiy/evil-merge-detector/internal/models"
)

// Reporter defines the interface for outputting scan results.
type Reporter interface {
	Report(w io.Writer, result *models.ScanResult) error
}
