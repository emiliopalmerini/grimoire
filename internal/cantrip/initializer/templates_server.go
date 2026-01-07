package initializer

import (
	"fmt"
	"slices"
)

func healthTemplate() string {
	return `package server

import (
	"encoding/json"
	"net/http"
)

type HealthResponse struct {
	Status string ` + "`json:\"status\"`" + `
}

func Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HealthResponse{Status: "ok"})
}
`
}

func httpServerTemplate(modulePath string, opts ProjectOptions) string {
	middlewareImports := ""
	middlewareUse := ""

	if opts.Type == "api" {
		middlewareUse = `	r.Use(middleware.CORS)
`
	}

	if opts.Type == "web" {
		middlewareImports = `
	"github.com/alexedwards/scs/v2"`
		middlewareUse = `	r.Use(middleware.Session(sessionManager))
	r.Use(middleware.CSRF(sessionManager))
`
	}

	webParam := ""
	webSessionManager := ""
	if opts.Type == "web" {
		webParam = ", sessionManager *scs.SessionManager"
		webSessionManager = `
	_ = sessionManager`
	}

	return fmt.Sprintf(`package server

import (
	"net/http"%s

	"github.com/go-chi/chi/v5"

	"%s/internal/app"
	"%s/internal/middleware"
)

func NewHTTPServer(cfg *app.Config%s) *http.Server {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.Logging)
	r.Use(middleware.Recovery)
%s%s
	r.Get("/health", Health)

	return &http.Server{
		Addr:    cfg.Addr,
		Handler: r,
	}
}
`, middlewareImports, modulePath, modulePath, webParam, middlewareUse, webSessionManager)
}

func grpcServerTemplate(modulePath string) string {
	return fmt.Sprintf(`package server

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func NewGRPCServer() *grpc.Server {
	srv := grpc.NewServer()

	healthSrv := health.NewServer()
	grpc_health_v1.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(srv)

	return srv
}
`)
}

func mainTemplate(modulePath string, opts ProjectOptions) string {
	imports := `	"context"
	"log"
	"os"
	"os/signal"
	"syscall"`

	if slices.Contains(opts.Transports, "http") {
		imports += `
	"net/http"`
	}

	if slices.Contains(opts.Transports, "grpc") {
		imports += `
	"net"`
	}

	if opts.Type == "web" {
		imports += `

	"github.com/alexedwards/scs/v2"`
	}

	imports += fmt.Sprintf(`

	"%s/internal/app"
	"%s/internal/server"`, modulePath, modulePath)

	serverStart := ""
	serverShutdown := ""

	if slices.Contains(opts.Transports, "http") {
		sessionManagerInit := ""
		sessionManagerParam := ""
		if opts.Type == "web" {
			sessionManagerInit = `
	sessionManager := scs.New()
`
			sessionManagerParam = ", sessionManager"
		}

		serverStart += fmt.Sprintf(`%s
	httpSrv := server.NewHTTPServer(cfg%s)
	go func() {
		log.Printf("http server listening on %%s", cfg.Addr)
		if err := httpSrv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()
`, sessionManagerInit, sessionManagerParam)

		serverShutdown += `
	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Printf("http shutdown error: %v", err)
	}`
	}

	if slices.Contains(opts.Transports, "grpc") {
		serverStart += `
	grpcSrv := server.NewGRPCServer()
	grpcLis, err := net.Listen("tcp", cfg.GRPCAddr)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		log.Printf("grpc server listening on %s", cfg.GRPCAddr)
		if err := grpcSrv.Serve(grpcLis); err != nil {
			log.Fatal(err)
		}
	}()
`
		serverShutdown += `
	grpcSrv.GracefulStop()`
	}

	return fmt.Sprintf(`package main

import (
%s
)

func main() {
	cfg, err := app.New()
	if err != nil {
		log.Fatal(err)
	}
%s
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
%s
}
`, imports, serverStart, serverShutdown)
}
