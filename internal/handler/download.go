package handler

import (
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/zhitoo/hls_converter/internal/models"
)

func (s *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	taskID := r.PathValue("task_id")

	t, errMsg := s.requireOwned(r, taskID)
	if errMsg != "" {
		status := http.StatusNotFound
		if errMsg == "access denied" || errMsg == "invalid task_id" {
			status = http.StatusForbidden
		}
		writeError(w, status, errMsg)
		return
	}

	if t.Status != models.StatusCompleted {
		writeError(w, http.StatusConflict, fmt.Sprintf("task is not completed (current status: %s)", t.Status))
		return
	}

	outputDir, err := s.stor.HLSOutputDir(t.UserID, t.TaskID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cannot resolve output directory")
		return
	}

	entries, err := os.ReadDir(outputDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "cannot read output directory")
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.zip"`, taskID))

	zw := zip.NewWriter(w)
	defer zw.Close()

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if err := addToZip(zw, filepath.Join(outputDir, e.Name()), e.Name()); err != nil {
			// Headers already sent; nothing to do but log.
			return
		}
	}
}

func addToZip(zw *zip.Writer, filePath, name string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	zf, err := zw.Create(name)
	if err != nil {
		return err
	}

	_, err = io.Copy(zf, f)
	return err
}
