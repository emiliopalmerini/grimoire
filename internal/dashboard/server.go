package dashboard

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/emiliopalmerini/grimorio/internal/metrics"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

//go:embed static/*
var staticFS embed.FS

type Server struct {
	tracker *metrics.SQLiteTracker
	port    int
	router  chi.Router
}

func NewServer(tracker *metrics.SQLiteTracker, port int) *Server {
	s := &Server{
		tracker: tracker,
		port:    port,
	}
	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	staticContent, _ := fs.Sub(staticFS, "static")
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(staticContent))))

	r.Get("/", s.handleDashboard)
	r.Get("/stats", s.handleStats)
	r.Get("/commands", s.handleCommands)
	r.Get("/models", s.handleModels)
	r.Get("/ai-activity", s.handleAIActivity)
	r.Get("/history", s.handleHistory)

	s.router = r
}

func (s *Server) Start() error {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: s.router,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-done
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
	}()

	fmt.Printf("Dashboard running at http://localhost:%d\n", s.port)
	fmt.Println("Press Ctrl+C to stop")

	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		return err
	}
	return nil
}
