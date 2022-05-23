package key

import (
	"encoding/json"
	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"
)

type ResponseItem struct {
	Status string `json:"status"`

	User    string `json:"user_service,omitempty"`
	Retro   string `json:"retro_service,omitempty"`
	Timer   string `json:"timer_service,omitempty"`
	Company string `json:"company_service,omitempty"`
	Billing string `json:"billing_service,omitempty"`
}

func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (k Key) CreateHandler(w http.ResponseWriter, r *http.Request) {
	user_id := r.Header.Get("X-User-ID")
	if user_id == "" {
		jsonResponse(w, http.StatusBadRequest, &ResponseItem{
			Status: "missing user-id",
		})
		return
	}

	keys := &ResponseItem{
		Status:  "ok",
		User:    k.generateServiceKey(25),
		Retro:   k.generateServiceKey(25),
		Timer:   k.generateServiceKey(25),
		Company: k.generateServiceKey(25),
		Billing: k.generateServiceKey(25),
	}

	if err := NewMongo(k.Config).Create(DataSet{
		UserID:    user_id,
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
		bugLog.Info(err)
		jsonResponse(w, http.StatusInternalServerError, &ResponseItem{
			Status: "internal error",
		})
		return
	}

	jsonResponse(w, http.StatusOK, keys)
}

func (k Key) GetHandler(w http.ResponseWriter, r *http.Request) {
	user_id := r.Header.Get("x-user-id")
	if user_id == "" {
		jsonResponse(w, http.StatusBadRequest, &ResponseItem{
			Status: "missing user-id",
		})
		return
	}

	keys, err := NewMongo(k.Config).Get(user_id)
	if err != nil {
		bugLog.Info(err)
		jsonResponse(w, http.StatusInternalServerError, &ResponseItem{
			Status: "internal error",
		})
		return
	}

	if keys == nil {
		jsonResponse(w, http.StatusNotFound, &ResponseItem{
			Status: "not found",
		})
		return
	}

	jsonResponse(w, http.StatusOK, &ResponseItem{
		Status:  "ok",
		User:    keys.Keys.UserService,
		Retro:   keys.Keys.RetroService,
		Timer:   keys.Keys.TimerService,
		Company: keys.Keys.CompanyService,
		Billing: keys.Keys.BillingService,
	})
}

func (k Key) CheckHandler(w http.ResponseWriter, r *http.Request) {
	user_id := r.Header.Get("x-user-id")
	if user_id == "" {
		jsonResponse(w, http.StatusBadRequest, &ResponseItem{
			Status: "missing user-id",
		})
		return
	}

	checkKey := chi.URLParam(r, "key")
	if checkKey == "" {
		jsonResponse(w, http.StatusBadRequest, &ResponseItem{
			Status: "missing key",
		})
		return
	}

	keys, err := NewMongo(k.Config).Get(user_id)
	if err != nil {
		bugLog.Info(err)
		jsonResponse(w, http.StatusInternalServerError, &ResponseItem{
			Status: "internal error",
		})
		return
	}

	if keys == nil {
		jsonResponse(w, http.StatusUnauthorized, &ResponseItem{
			Status: "not allowed",
		})
		return
	}

	userKey := keys.Keys.UserService
	retroKey := keys.Keys.RetroService
	timerKey := keys.Keys.TimerService
	companyKey := keys.Keys.CompanyService
	billingKey := keys.Keys.BillingService

	if checkKey == userKey || checkKey == retroKey || checkKey == timerKey || checkKey == companyKey || checkKey == billingKey {
		jsonResponse(w, http.StatusOK, &ResponseItem{
			Status: "ok",
		})
		return
	}

	jsonResponse(w, http.StatusUnauthorized, &ResponseItem{
		Status: "not allowed",
	})
}
