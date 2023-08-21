package service

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/bugfixes/go-bugfixes/logs"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	gc "github.com/keloran/go-config"
	"github.com/keloran/go-healthcheck"
	"github.com/keloran/go-probe"
	"github.com/todo-lists-app/id-checker/internal/checker"
	pb "github.com/todo-lists-app/protobufs/generated/id_checker/v1"
	"google.golang.org/grpc"
)

// Service is the service
type Service struct {
	Config *gc.Config
}

// Start the service
func (s *Service) Start() error {
	errChan := make(chan error)

	go startGRPC(s.Config.Local.GRPCPort, errChan, s.Config)
	go startHTTP(s.Config.Local.HTTPPort, errChan)

	return <-errChan
}

func startGRPC(port int, errChan chan error, cfg *gc.Config) {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		errChan <- logs.Errorf("failed to listen: %v", err)
	}
	gs := grpc.NewServer()
	logs.Local().Infof("starting grpc on %s", lis.Addr().String())
	pb.RegisterIdCheckerServiceServer(gs, &checker.Server{
		Config: cfg,
	})
	if err := gs.Serve(lis); err != nil {
		errChan <- logs.Errorf("failed to serve: %v", err)
	}
}

func startHTTP(port int, errChan chan error) {
	r := chi.NewRouter()
	r.Use(middleware.Heartbeat("/ping"))
	r.Get("/health", healthcheck.HTTP)
	r.Get("/probe", probe.HTTP)

	srv := http.Server{
		Addr:              fmt.Sprintf(":%d", port),
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       15 * time.Second,
		ReadTimeout:       15 * time.Second,
	}
	logs.Local().Infof("starting http on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		errChan <- logs.Errorf("failed to serve: %v", err)
	}
}
