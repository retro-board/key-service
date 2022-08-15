package key

import (
	"context"
	"time"

	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/retro-board/key-service/internal/config"
	pb "github.com/retro-board/protos/generated/key/v1"
)

type Server struct {
	pb.UnimplementedKeyServiceServer
	Config *config.Config
}

func (s *Server) Create(c context.Context, r *pb.CreateRequest) (*pb.KeyResponse, error) {
	if r.UserId == "" {
		status := "missing user-id"
		bugLog.Info(status)
		return &pb.KeyResponse{
			Status: status,
		}, nil
	}

	if r.ServiceKey == "" {
		status := "missing service-key"
		bugLog.Info(status)
		return &pb.KeyResponse{
			Status: status,
		}, nil
	}

	k := NewKey(s.Config)
	if !k.ValidateServiceKey(r.ServiceKey) {
		status := "invalid service key"
		bugLog.Info(status)
		return &pb.KeyResponse{
			Status: status,
		}, nil
	}

	keys, err := k.GetKeys(25)
	if err != nil {
		bugLog.Info(err)
		status := "internal error, 1"
		return &pb.KeyResponse{
			Status: status,
		}, nil
	}

	if err := NewMongo(k.Config).Create(DataSet{
		UserID:    r.UserId,
		Generated: time.Now().Unix(),
		Keys: struct {
			UserService        string `json:"user_service" bson:"user_service"`
			RetroService       string `json:"retro_service" bson:"retro_service"`
			TimerService       string `json:"timer_service" bson:"timer_service"`
			CompanyService     string `json:"company_service" bson:"company_service"`
			BillingService     string `json:"billing_service" bson:"billing_service"`
			PermissionsService string `json:"permissions_service" bson:"permissions_service"`
		}{
			UserService:        keys.User,
			RetroService:       keys.Retro,
			TimerService:       keys.Timer,
			CompanyService:     keys.Company,
			BillingService:     keys.Billing,
			PermissionsService: keys.Permissions,
		},
	}); err != nil {
		bugLog.Info(err)
		status := "internal error, 2"
		return &pb.KeyResponse{
			Status: status,
		}, nil
	}

	return &pb.KeyResponse{
		User:    keys.User,
		Retro:   keys.Retro,
		Timer:   keys.Timer,
		Company: keys.Company,
		Billing: keys.Billing,
	}, nil
}

func (s *Server) Get(c context.Context, r *pb.GetRequest) (*pb.KeyResponse, error) {
	if r.UserId == "" {
		status := "missing user-id"
		bugLog.Info(status)
		return &pb.KeyResponse{
			Status: status,
		}, nil
	}

	if r.ServiceKey == "" {
		status := "missing service-key"
		bugLog.Info(status)
		return &pb.KeyResponse{
			Status: status,
		}, nil
	}

	k := NewKey(s.Config)
	if !k.ValidateServiceKey(r.ServiceKey) {
		status := "invalid service key"
		bugLog.Info(status)
		return &pb.KeyResponse{
			Status: status,
		}, nil
	}

	keys, err := NewMongo(k.Config).Get(r.UserId)
	if err != nil {
		bugLog.Info(err)
		status := "internal error, 3"
		return &pb.KeyResponse{
			Status: status,
		}, nil
	}

	if keys == nil {
		status := "user not found"
		bugLog.Info("no keys or expired for user")
		return &pb.KeyResponse{
			Status: status,
		}, nil
	}

	return &pb.KeyResponse{
		User:        keys.Keys.UserService,
		Retro:       keys.Keys.RetroService,
		Timer:       keys.Keys.TimerService,
		Company:     keys.Keys.CompanyService,
		Billing:     keys.Keys.BillingService,
		Permissions: keys.Keys.PermissionsService,
	}, nil
}

//nolint:gocyclo
func (s *Server) Validate(c context.Context, r *pb.ValidateRequest) (*pb.ValidResponse, error) {
	if r.UserId == "" {
		status := "missing user-id"
		bugLog.Info(status)
		return &pb.ValidResponse{
			Status: status,
		}, nil
	}

	if r.ServiceKey == "" {
		status := "missing service-key"
		bugLog.Info(status)
		return &pb.ValidResponse{
			Status: status,
		}, nil
	}

	if r.CheckKey == "" {
		status := "missing check-key"
		bugLog.Info(status)
		return &pb.ValidResponse{
			Status: status,
		}, nil
	}

	k := NewKey(s.Config)
	if !k.ValidateServiceKey(r.ServiceKey) {
		status := "invalid service key"
		bugLog.Info(status)
		return &pb.ValidResponse{
			Valid:  false,
			Status: status,
		}, nil
	}

	if s.Config.Local.Development {
		return &pb.ValidResponse{
			Valid: true,
		}, nil
	}

	keys, err := NewMongo(k.Config).Get(r.UserId)
	if err != nil {
		status := "internal error, 4"
		bugLog.Info(err)
		return &pb.ValidResponse{
			Valid:  false,
			Status: status,
		}, nil
	}

	if r.CheckKey == keys.Keys.UserService ||
		r.CheckKey == keys.Keys.RetroService ||
		r.CheckKey == keys.Keys.TimerService ||
		r.CheckKey == keys.Keys.CompanyService ||
		r.CheckKey == keys.Keys.BillingService ||
		r.CheckKey == keys.Keys.PermissionsService {
		return &pb.ValidResponse{
			Valid: true,
		}, nil
	}

	return &pb.ValidResponse{
		Valid: false,
	}, nil
}
