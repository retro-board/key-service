package key

import (
	"context"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/retro-board/key-service/internal/config"
	pb "github.com/retro-board/protos/generated/key/v1"
	"time"
)

type Server struct {
	pb.UnimplementedKeyServiceServer
	Config *config.Config
}

func (s *Server) Create(c context.Context, r *pb.CreateRequest) (*pb.KeyResponse, error) {
	if r.UserId == "" {
		bugLog.Info("missing user-id")
		return &pb.KeyResponse{
			Status: "missing user-id",
		}, nil
	}

	if r.ServiceKey == "" {
		bugLog.Info("missing service-key")
		return &pb.KeyResponse{
			Status: "missing service-key",
		}, nil
	}

	k := NewKey(s.Config)
	if !k.ValidateServiceKey(r.ServiceKey) {
		bugLog.Info("invalid service key")
		return &pb.KeyResponse{
			Status: "invalid service key",
		}, nil
	}

	keys, err := k.GetKeys(25)
	if err != nil {
		bugLog.Info(err)
		return &pb.KeyResponse{
			Status: "internal error, 1",
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
		return &pb.KeyResponse{
			Status: "internal error, 2",
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
		bugLog.Info("missing user-id")
		return &pb.KeyResponse{
			Status: "missing user-id",
		}, nil
	}

	if r.ServiceKey == "" {
		bugLog.Info("missing service-key")
		return &pb.KeyResponse{
			Status: "missing service-key",
		}, nil
	}

	k := NewKey(s.Config)
	if !k.ValidateServiceKey(r.ServiceKey) {
		bugLog.Info("invalid service key")
		return &pb.KeyResponse{
			Status: "invalid service key",
		}, nil
	}

	keys, err := NewMongo(k.Config).Get(r.UserId)
	if err != nil {
		bugLog.Info(err)
		return &pb.KeyResponse{
			Status: "internal error",
		}, nil
	}

	if keys == nil {
		bugLog.Info("no keys or expired for user")
		return &pb.KeyResponse{
			Status: "user not found",
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

func (s *Server) Validate(c context.Context, r *pb.ValidateRequest) (*pb.ValidResponse, error) {
	if r.UserId == "" {
		bugLog.Info("missing user-id")
		return &pb.ValidResponse{
			Status: "missing user-id",
		}, nil
	}

	if r.ServiceKey == "" {
		bugLog.Info("missing service-key")
		return &pb.ValidResponse{}, nil
	}

	if r.CheckKey == "" {
		bugLog.Info("missing check-key")
		return &pb.ValidResponse{
			Status: "missing check-key",
		}, nil
	}

	k := NewKey(s.Config)
	if !k.ValidateServiceKey(r.ServiceKey) {
		bugLog.Info("invalid service key")
		return &pb.ValidResponse{
			Valid:  false,
			Status: "invalid service key",
		}, nil
	}

	keys, err := NewMongo(k.Config).Get(r.UserId)
	if err != nil {
		return &pb.ValidResponse{
			Valid:  false,
			Status: "internal error",
		}, nil
	}

	if r.CheckKey == keys.Keys.UserService ||
		r.CheckKey == keys.Keys.RetroService ||
		r.CheckKey == keys.Keys.TimerService ||
		r.CheckKey == keys.Keys.CompanyService ||
		r.CheckKey == keys.Keys.BillingService {
		return &pb.ValidResponse{
			Valid: true,
		}, nil
	}

	return &pb.ValidResponse{
		Valid: false,
	}, nil
}
