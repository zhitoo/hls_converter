package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/zhitoo/hls_converter/internal/auth"
	"github.com/zhitoo/hls_converter/internal/cleanup"
	"github.com/zhitoo/hls_converter/internal/converter"
	"github.com/zhitoo/hls_converter/internal/handler"
	"github.com/zhitoo/hls_converter/internal/queue"
	"github.com/zhitoo/hls_converter/internal/storage"
	"github.com/zhitoo/hls_converter/internal/task"
	"github.com/zhitoo/hls_converter/internal/tasklog"
)

const (
	addr       = ":8080"
	workerPool = 3
	queueSize  = 1000
)

func main() {
	projectRoot := "."

	userRepo, err := auth.NewFileRepository(filepath.Join(projectRoot, "users.json"))
	if err != nil {
		log.Fatalf("loading users: %v", err)
	}

	stor, err := storage.New(filepath.Join(projectRoot, "storage"))
	if err != nil {
		log.Fatalf("init storage: %v", err)
	}

	tasksDir := filepath.Join(projectRoot, "tasks")
	if err := os.MkdirAll(tasksDir, 0o755); err != nil {
		log.Fatalf("creating tasks dir: %v", err)
	}

	taskRepo := task.NewFileRepository(tasksDir)
	taskMgr := task.NewManager(taskRepo)

	logFactory := tasklog.NewFactory(filepath.Join(projectRoot, "storage", "logs"))

	q := queue.New(queueSize)
	conv := converter.NewFFmpeg()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cleaner := cleanup.New(taskRepo, stor)
	cleaner.DrainStale()
	go cleaner.Run(ctx, time.Minute)

	for i := 0; i < workerPool; i++ {
		w := queue.NewWorker(taskRepo, conv, logFactory, stor)
		go w.Run(ctx, q.Channel())
	}

	authMW := auth.Middleware(userRepo)
	srv := handler.NewServer(taskMgr, taskRepo, stor, q, authMW)

	httpServer := &http.Server{
		Addr:    addr,
		Handler: srv.Routes(),
	}

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
		<-sig
		log.Println("shutting down...")
		cancel()
		_ = httpServer.Shutdown(context.Background())
	}()

	log.Printf("server listening on %s", addr)
	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}
