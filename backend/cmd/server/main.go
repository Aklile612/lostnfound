package main

import (
	"context"
	"log/slog"
	"lostandfound/internal/ai"
	"lostandfound/internal/config"
	"lostandfound/internal/handler"
	"lostandfound/internal/port"
	"lostandfound/internal/repository"
	"lostandfound/internal/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}
	if err := handler.EnsureUploadDir(cfg.UploadDir); err != nil {
		slog.Error("uploads", "err", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	db, err := repository.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("database", "err", err)
		os.Exit(1)
	}
	defer db.Close()

	var groq port.TextMatcher
	if cfg.GroqAPIKey != "" {
		groq = ai.NewGroq(cfg.GroqAPIKey, cfg.GroqModel)
	}
	var gemini port.VisionMatcher
	if cfg.GeminiAPIKey != "" {
		gemini = ai.NewGemini(cfg.GeminiAPIKey, cfg.GeminiModel, cfg.UploadDir)
	}

	svc := service.New(db, db, groq, gemini, cfg.MatchWindow, cfg.MatchThreshold)
	h := handler.New(svc, cfg.UploadDir)
	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler.Router(h, handler.NormalizeOrigin(cfg.CORSOrigin), cfg.UploadDir),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("listening", "port", cfg.Port, "groq", cfg.GroqAPIKey != "", "gemini", cfg.GeminiAPIKey != "")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	shutdown, stop := context.WithTimeout(context.Background(), 8*time.Second)
	defer stop()
	_ = server.Shutdown(shutdown)
}
