package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"net"

	"github.com/test/grpc/internal/app"
	"github.com/test/grpc/internal/server"
)

func main() {
	cfg, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

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

	grpcSrv.GracefulStop()
}
