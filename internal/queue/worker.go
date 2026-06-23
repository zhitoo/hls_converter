package queue

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/zhitoo/hls_converter/internal/converter"
	"github.com/zhitoo/hls_converter/internal/models"
	"github.com/zhitoo/hls_converter/internal/storage"
	"github.com/zhitoo/hls_converter/internal/task"
	"github.com/zhitoo/hls_converter/internal/tasklog"
)

const maxRetries = 4

// Worker consumes tasks from a channel, runs conversion with retry logic,
// and keeps the task state file up-to-date throughout.
type Worker struct {
	taskRepo   task.Repository
	converter  converter.Converter
	logFactory tasklog.Factory
	storage    storage.Storage
}

func NewWorker(
	taskRepo task.Repository,
	conv converter.Converter,
	logFactory tasklog.Factory,
	stor storage.Storage,
) *Worker {
	return &Worker{
		taskRepo:   taskRepo,
		converter:  conv,
		logFactory: logFactory,
		storage:    stor,
	}
}

// Run processes tasks from jobs until ctx is cancelled.
func (w *Worker) Run(ctx context.Context, jobs <-chan *models.Task) {
	for {
		select {
		case <-ctx.Done():
			return
		case t := <-jobs:
			w.process(ctx, t)
		}
	}
}

func (w *Worker) process(ctx context.Context, t *models.Task) {
	log.Printf("[task:%s] worker picked up task", t.TaskID)

	logger, err := w.logFactory.New(t.TaskID)
	if err != nil {
		log.Printf("[task:%s] cannot open log file: %v", t.TaskID, err)
		w.fail(t, "logger unavailable")
		return
	}
	defer logger.Close()

	baseDir, err := w.storage.HLSOutputDir(t.UserID, t.TaskID)
	if err != nil {
		log.Printf("[task:%s] cannot create output directory: %v", t.TaskID, err)
		w.fail(t, "storage error")
		return
	}

	resolutions := t.Config.Resolutions
	log.Printf("[task:%s] config resolutions: %v", t.TaskID, resolutions)
	if len(resolutions) == 0 {
		resolutions = []int{0}
		log.Printf("[task:%s] no resolutions configured, using original only", t.TaskID)
	}
	total := len(resolutions)

	for idx, height := range resolutions {
		outputDir := baseDir
		stepLabel := "Converting"
		if height > 0 {
			outputDir = filepath.Join(baseDir, fmt.Sprintf("%dp", height))
			if err := os.MkdirAll(outputDir, 0o755); err != nil {
				log.Printf("[task:%s] cannot create output dir for %dp: %v", t.TaskID, height, err)
				w.fail(t, fmt.Sprintf("storage error for %dp", height))
				return
			}
			stepLabel = fmt.Sprintf("Converting %dp", height)
		}

		cfg := t.Config
		if height > 0 {
			cfg.Resolutions = []int{height}
		} else {
			cfg.Resolutions = nil
		}

		var convErr error
		for attempt := 0; attempt <= maxRetries; attempt++ {
			if attempt == 0 {
				t.CurrentStep = stepLabel
			} else {
				t.CurrentStep = fmt.Sprintf("%s – retry %d/%d", stepLabel, attempt, maxRetries)
				log.Printf("[task:%s] retry attempt %d/%d for %s", t.TaskID, attempt, maxRetries, stepLabel)
				time.Sleep(time.Duration(attempt) * 2 * time.Second)
			}

			t.Status = models.StatusProcessing
			t.RetryCount = attempt
			t.UpdatedAt = time.Now()
			_ = w.taskRepo.Save(t)

			convErr = w.converter.Convert(ctx, cfg, outputDir, logger, func(pct int, step string) {
				overall := (idx*100 + pct) / total
				t.Progress = overall
				t.CurrentStep = fmt.Sprintf("%s (%d/%d)", stepLabel, idx+1, total)
				t.UpdatedAt = time.Now()
				_ = w.taskRepo.Save(t)
			})

			if convErr == nil {
				log.Printf("[task:%s] %s completed", t.TaskID, stepLabel)
				break
			}
			log.Printf("[task:%s] %s attempt %d failed: %v", t.TaskID, stepLabel, attempt+1, convErr)
		}

		if convErr != nil {
			log.Printf("[task:%s] %s failed after %d attempts", t.TaskID, stepLabel, maxRetries+1)
			w.fail(t, fmt.Sprintf("%s failed after %d attempts", stepLabel, maxRetries+1))
			return
		}
	}

	log.Printf("[task:%s] writing master playlist for resolutions: %v", t.TaskID, resolutions)
	if err := writeMasterPlaylist(baseDir, resolutions); err != nil {
		log.Printf("[task:%s] WARNING: master playlist write failed: %v", t.TaskID, err)
	} else {
		log.Printf("[task:%s] SUCCESS: master.m3u8 created at %s", t.TaskID, filepath.Join(baseDir, "master.m3u8"))
	}

	t.Status = models.StatusCompleted
	t.Progress = 100
	t.CurrentStep = "Completed"
	t.UpdatedAt = time.Now()
	_ = w.taskRepo.Save(t)
	log.Printf("[task:%s] all conversions completed successfully", t.TaskID)
}

var resolutionMeta = map[int]struct {
	bandwidth  int
	resolution string
}{
	240:  {400_000, "426x240"},
	360:  {800_000, "640x360"},
	480:  {1_400_000, "854x480"},
	720:  {2_800_000, "1280x720"},
	1080: {5_000_000, "1920x1080"},
}

func writeMasterPlaylist(baseDir string, heights []int) error {
	var b strings.Builder
	b.WriteString("#EXTM3U\n#EXT-X-VERSION:3\n\n")
	for _, h := range heights {
		meta, ok := resolutionMeta[h]
		if !ok {
			meta.bandwidth = h * 4000
			meta.resolution = fmt.Sprintf("?x%d", h)
		}
		fmt.Fprintf(&b, "#EXT-X-STREAM-INF:BANDWIDTH=%d,RESOLUTION=%s\n%dp/output.m3u8\n\n",
			meta.bandwidth, meta.resolution, h)
	}
	return os.WriteFile(filepath.Join(baseDir, "master.m3u8"), []byte(b.String()), 0o644)
}

func (w *Worker) fail(t *models.Task, reason string) {
	log.Printf("[task:%s] FAILED: %s", t.TaskID, reason)
	t.Status = models.StatusFailed
	t.CurrentStep = reason
	t.UpdatedAt = time.Now()
	_ = w.taskRepo.Save(t)
}
