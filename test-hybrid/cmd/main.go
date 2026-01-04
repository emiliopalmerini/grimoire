package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"net/http"
	"net"

	"github.com/test/hybrid/internal/app"
	"github.com/test/hybrid/internal/server"
)

func main() {
	cfg, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	httpSrv := server.NewHTTPServer(cfg)
	go func() {
		log.Printf("http server listening on %s", cfg.Addr)
		if err := httpSrv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

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

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := httpSrv.Shutdown(ctx); err != nil {
		log.Printf("http shutdown error: %v", err)
	}
	grpcSrv.GracefulStop()
}
