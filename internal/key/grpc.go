package key

import (
	"context"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/retro-board/key-service/internal/config"
	pb "github.com/retro-board/protos/generated/v1"
	"time"
)

type Server struct {
	pb.UnimplementedKeyServiceServer
	Config *config.Config
}

func (s *Server) Create(c context.Context, r *pb.CreateRequest) (*pb.KeyResponse, error) {
	if r.UserId == "" {
		return nil, bugLog.Errorf("missing user-id")
	}

	if r.ServiceKey == "" {
		return nil, bugLog.Errorf("missing service-key")
	}

	k := NewKey(s.Config)
	if !k.ValidateServiceKey(r.ServiceKey) {
		return nil, bugLog.Errorf("invalid service key")
	}

	keys, err := k.GetKeys(25)
	if err != nil {
		return nil, bugLog.Errorf("internal error")
	}

	if err := NewMongo(k.Config).Create(DataSet{
		UserID:    r.UserId,
		Generated: time.Now().Unix(),
		Keys: struct {
			UserService    string `json:"user_service" bson:"user_service"`
			RetroService   string `json:"retro_service" bson:"retro_service"`
			TimerService   string `json:"timer_service" bson:"timer_service"`
			CompanyService string `json:"company_service" bson:"company_service"`
			BillingService string `json:"billing_service" bson:"billing_service"`
		}{
			UserService:    keys.User,
			RetroService:   keys.Retro,
			TimerService:   keys.Timer,
			CompanyService: keys.Company,
			BillingService: keys.Billing,
		},
	}); err != nil {
		return nil, bugLog.Errorf("internal error")
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
		return nil, bugLog.Errorf("missing user-id")
	}

	if r.ServiceKey == "" {
		return nil, bugLog.Errorf("missing service-key")
	}

	k := NewKey(s.Config)
	if !k.ValidateServiceKey(r.ServiceKey) {
		return nil, bugLog.Errorf("invalid service key")
	}

	keys, err := NewMongo(k.Config).Get(r.UserId)
	if err != nil {
		bugLog.Info(err)
		return nil, bugLog.Errorf("internal error")
	}

	if keys == nil {
		return nil, bugLog.Errorf("no keys or expired for user")
	}

	return &pb.KeyResponse{
		User:    keys.Keys.UserService,
		Retro:   keys.Keys.RetroService,
		Timer:   keys.Keys.TimerService,
		Company: keys.Keys.CompanyService,
		Billing: keys.Keys.BillingService,
	}, nil
}

func (s *Server) Validate(c context.Context, r *pb.ValidateRequest) (*pb.ValidResponse, error) {
	if r.UserId == "" {
		return nil, bugLog.Errorf("missing user-id")
	}

	if r.ServiceKey == "" {
		return nil, bugLog.Errorf("missing service-key")
	}

	if r.CheckKey == "" {
		return nil, bugLog.Errorf("missing check-key")
	}

	k := NewKey(s.Config)
	if !k.ValidateServiceKey(r.ServiceKey) {
		return nil, bugLog.Errorf("invalid service key")
	}

	keys, err := NewMongo(k.Config).Get(r.UserId)
	if err != nil {
		return &pb.ValidResponse{
			Valid: false,
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
