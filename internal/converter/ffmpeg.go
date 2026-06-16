package converter

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"github.com/zhitoo/hls_converter/internal/models"
)

var (
	durationRe = regexp.MustCompile(`Duration:\s*(\d+):(\d+):(\d+\.\d+)`)
	timeRe     = regexp.MustCompile(`time=(\d+):(\d+):(\d+\.\d+)`)
)

type ffmpegConverter struct{}

func NewFFmpeg() Converter {
	return &ffmpegConverter{}
}

func (c *ffmpegConverter) Convert(
	ctx context.Context,
	cfg models.HLSConfig,
	outputDir string,
	log io.Writer,
	onProgress func(pct int, step string),
) error {
	m3u8Path := filepath.Join(outputDir, "output.m3u8")
	segPattern := filepath.Join(outputDir, "segment_%03d.ts")

	var height int
	if len(cfg.Resolutions) > 0 {
		height = cfg.Resolutions[0]
	}
	args := buildArgs(cfg, height, segPattern, m3u8Path)
	cmd := exec.CommandContext(ctx, "ffmpeg", args...)

	// Pipe stderr so we can tee it to the log file and parse progress.
	pr, pw := io.Pipe()
	cmd.Stderr = pw

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("ffmpeg start: %w", err)
	}

	// Parse progress in a goroutine while tee-ing raw bytes to the log writer.
	done := make(chan struct{})
	go func() {
		defer close(done)
		var totalSecs float64

		scanner := bufio.NewScanner(io.TeeReader(pr, log))
		scanner.Split(splitOnNewlineOrCR)

		for scanner.Scan() {
			line := scanner.Text()

			if totalSecs == 0 {
				if m := durationRe.FindStringSubmatch(line); len(m) == 4 {
					totalSecs = toSeconds(m[1], m[2], m[3])
				}
			}

			if totalSecs > 0 {
				if m := timeRe.FindStringSubmatch(line); len(m) == 4 {
					cur := toSeconds(m[1], m[2], m[3])
					pct := int(cur / totalSecs * 100)
					if pct > 99 {
						pct = 99
					}
					if onProgress != nil {
						onProgress(pct, "Converting")
					}
				}
			}
		}
	}()

	err := cmd.Wait()
	pw.Close()
	<-done

	return err
}

func buildArgs(cfg models.HLSConfig, height int, segPattern, m3u8Path string) []string {
	args := []string{"-y", "-i", cfg.VideoURL}

	if height > 0 {
		args = append(args, "-vf", fmt.Sprintf("scale=-2:%d", height))
		args = append(args, "-c:v", "libx264")
	} else {
		args = append(args, "-c:v", "copy")
	}

	if cfg.AudioChannels > 0 {
		args = append(args, "-ac", strconv.Itoa(cfg.AudioChannels))
	}

	args = append(args,
		"-c:a", "aac",
		"-hls_time", strconv.Itoa(cfg.ChunkDuration),
		"-hls_playlist_type", "vod",
		"-hls_segment_filename", segPattern,
		m3u8Path,
	)
	return args
}

func toSeconds(h, m, s string) float64 {
	hours, _ := strconv.ParseFloat(h, 64)
	mins, _ := strconv.ParseFloat(m, 64)
	secs, _ := strconv.ParseFloat(s, 64)
	return hours*3600 + mins*60 + secs
}

// splitOnNewlineOrCR is a bufio.SplitFunc that splits on \n or \r so that
// FFmpeg's carriage-return progress lines are treated as separate tokens.
func splitOnNewlineOrCR(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	for i, c := range data {
		if c == '\n' || c == '\r' {
			return i + 1, data[:i], nil
		}
	}
	if atEOF {
		return len(data), data, nil
	}
	return 0, nil, nil
}
