package reporter

import (
	"io"

	"github.com/evilmerge-dev/evil-merge-detector/internal/models"
)

// Reporter defines the interface for outputting scan results.
type Reporter interface {
	Report(w io.Writer, result *models.ScanResult) error
}
