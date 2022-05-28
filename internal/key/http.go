package key

import (
	"encoding/json"
	"net/http"
	"time"

	bugLog "github.com/bugfixes/go-bugfixes/logs"
	"github.com/go-chi/chi/v5"
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
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		jsonResponse(w, http.StatusBadRequest, &ResponseItem{
			Status: "missing user-id",
		})
		return
	}

	vaultKey := r.Header.Get("X-Service-Key")
	if vaultKey == "" {
		jsonResponse(w, http.StatusBadRequest, &ResponseItem{
			Status: "missing vault-key",
		})
		return
	}

	keys, err := k.getKeys(25)
	if err != nil {
		bugLog.Info(err)
		jsonResponse(w, http.StatusInternalServerError, &ResponseItem{
			Status: "internal error",
		})
		return
	}

	if err := NewMongo(k.Config).Create(DataSet{
		UserID:    userID,
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
	userID := r.Header.Get("x-user-id")
	if userID == "" {
		jsonResponse(w, http.StatusBadRequest, &ResponseItem{
			Status: "missing user-id",
		})
		return
	}

	keys, err := NewMongo(k.Config).Get(userID)
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
	userID := r.Header.Get("x-user-id")
	if userID == "" {
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

	keys, err := NewMongo(k.Config).Get(userID)
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
