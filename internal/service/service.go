package service

import (
	"fmt"
	"net"
	"net/http"

	bugLog "github.com/bugfixes/go-bugfixes/logs"
	bugMiddleware "github.com/bugfixes/go-bugfixes/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/keloran/go-healthcheck"
	"github.com/keloran/go-probe"
	"github.com/retro-board/key-service/internal/config"
	"github.com/retro-board/key-service/internal/key"
	pb "github.com/retro-board/protos/generated/key/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	kitlog "github.com/go-kit/log"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/kit"
)

type Service struct {
	Config *config.Config
}

func (s *Service) Start() error {
	errChan := make(chan error)
	go startGRPC(s.Config.GRPCPort, errChan, s.Config)
	go startHTTP(s.Config.HTTPPort, errChan, s.Config.Development)

	return <-errChan
}

func startGRPC(port int, errChan chan error, config *config.Config) {
	kOpts := []kit.Option{
		kit.WithDecider(func(methodFullName string, err error) bool {
			if err != nil {
				bugLog.Local().Infof("%s: %+v", methodFullName, err)
				return false
			}
			return true
		}),
	}
	opts := []grpc.ServerOption{
		grpc_middleware.WithStreamServerChain(
			kit.StreamServerInterceptor(kitlog.NewNopLogger(), kOpts...),
		),
		grpc_middleware.WithUnaryServerChain(
			kit.UnaryServerInterceptor(kitlog.NewNopLogger(), kOpts...),
		),
	}

	p := fmt.Sprintf(":%d", port)
	bugLog.Local().Infof("Starting Key GRPC: %s", p)
	lis, err := net.Listen("tcp", p)
	if err != nil {
		errChan <- bugLog.Errorf("failed to listen: %v", err)
	}
	gs := grpc.NewServer(opts...)
	reflection.Register(gs)
	pb.RegisterKeyServiceServer(gs, &key.Server{
		Config: config,
	})
	if err := gs.Serve(lis); err != nil {
		errChan <- bugLog.Errorf("failed to start grpc: %v", err)
	}
}

func startHTTP(port int, errChan chan error, development bool) {
	p := fmt.Sprintf(":%d", port)
	bugLog.Local().Infof("Starting Key HTTP: %s", p)

	allowedOrigins := []string{
		"http://localhost:8080",
		"https://retro-board.it",
		"https://*.retro-board.it",
	}
	if development {
		allowedOrigins = append(allowedOrigins, "http://*")
	}

	c := cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-User-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	})

	r := chi.NewRouter()
	r.Use(middleware.Heartbeat("/ping"))
	r.Use(middleware.RequestID)
	r.Use(c.Handler)
	r.Use(bugMiddleware.BugFixes)
	r.Get("/health", healthcheck.HTTP)
	r.Get("/probe", probe.HTTP)
	if err := http.ListenAndServe(p, r); err != nil {
		errChan <- bugLog.Errorf("port failed: %+v", err)
	}
}
