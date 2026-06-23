package converter

import (
	"context"
	"io"

	"github.com/zhitoo/hls_converter/internal/models"
)

// Converter converts a video to HLS format.
// onProgress is called periodically with the completion percentage (0–100) and current step label.
type Converter interface {
	Convert(ctx context.Context, cfg models.HLSConfig, outputDir string, log io.Writer, onProgress func(pct int, step string)) error
}
