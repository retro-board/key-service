package key

import (
	"encoding/json"
	"net/http"
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
	user_id := r.Header.Get("user_id")
	if user_id == "" {
		jsonResponse(w, http.StatusBadRequest, &ResponseItem{
			Status: "missing user_id",
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
	jsonResponse(w, http.StatusOK, keys)
}

func (k Key) CheckHandler(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusNotFound, &ResponseItem{
		Status: "not found",
	})
}
