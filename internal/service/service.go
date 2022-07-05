package service

import (
	"fmt"
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
	"net"
	"net/http"
)

type Service struct {
	Config *config.Config
}

func (s *Service) Start() error {
	errChan := make(chan error)
	go func() {
		port := fmt.Sprintf(":%d", s.Config.GRPCPort)
		bugLog.Local().Infof("Starting Key GRPC: %s", port)
		lis, err := net.Listen("tcp", port)
		if err != nil {
			errChan <- bugLog.Errorf("failed to listen: %v", err)
		}
		gs := grpc.NewServer()
		reflection.Register(gs)
		pb.RegisterKeyServiceServer(gs, &key.Server{
			Config: s.Config,
		})
		if err := gs.Serve(lis); err != nil {
			errChan <- bugLog.Errorf("failed to start grpc: %v", err)
		}
	}()

	go func() {
		port := fmt.Sprintf(":%d", s.Config.HTTPPort)
		bugLog.Local().Infof("Starting Key HTTP: %s", port)

		allowedOrigins := []string{
			"http://localhost:8080",
			"https://retro-board.it",
			"https://*.retro-board.it",
		}
		if s.Config.Development {
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
		if err := http.ListenAndServe(port, r); err != nil {
			errChan <- bugLog.Errorf("port failed: %+v", err)
		}
	}()

	return <-errChan
}
