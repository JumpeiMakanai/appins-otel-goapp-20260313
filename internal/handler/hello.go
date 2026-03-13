package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"

	"go-app/internal/service"
)

var profileClient = service.NewProfileClient(getEnv("PROFILE_SERVICE_URL", "http://localhost:8081"))

func HelloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	shouldFail := r.URL.Query().Get("fail") == "true"

	err := service.ExecuteBusiness(r.Context(), shouldFail)
	if err != nil {
		http.Error(w, fmt.Sprintf("error: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("hello"))
}

func HealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	shouldFail := r.URL.Query().Get("fail") == "true"

	err := service.ExecuteBusiness(r.Context(), shouldFail)
	if err != nil {
		http.Error(w, fmt.Sprintf("error: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status": "ok",
	})
}

func UsersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w, http.MethodGet)
		return
	}

	shouldFail := r.URL.Query().Get("fail") == "true"

	err := service.ExecuteBusiness(r.Context(), shouldFail)
	if err != nil {
		http.Error(w, fmt.Sprintf("error: %v", err), http.StatusInternalServerError)
		return
	}

	id := 1
	if rawID := r.URL.Query().Get("id"); rawID != "" {
		parsed, err := strconv.Atoi(rawID)
		if err != nil {
			http.Error(w, "error: invalid id", http.StatusBadRequest)
			return
		}
		id = parsed
	}

	profile, err := profileClient.GetProfile(r.Context(), id)
	if err != nil {
		http.Error(w, fmt.Sprintf("error: failed to get profile: %v", err), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user": profile,
	})
}

func EchoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	shouldFail := r.URL.Query().Get("fail") == "true"

	err := service.ExecuteBusiness(r.Context(), shouldFail)
	if err != nil {
		http.Error(w, fmt.Sprintf("error: %v", err), http.StatusInternalServerError)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error: failed to read request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "received request body",
		"body":    string(body),
	})
}

type CreateItemRequest struct {
	Name  string `json:"name"`
	Price int    `json:"price"`
}

func CreateItemHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w, http.MethodPost)
		return
	}

	shouldFail := r.URL.Query().Get("fail") == "true"

	err := service.ExecuteBusiness(r.Context(), shouldFail)
	if err != nil {
		http.Error(w, fmt.Sprintf("error: %v", err), http.StatusInternalServerError)
		return
	}

	defer r.Body.Close()

	var req CreateItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "error: invalid json body", http.StatusBadRequest)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"message": "item created",
		"item": map[string]any{
			"id":    1,
			"name":  req.Name,
			"price": req.Price,
		},
	})
}

func writeMethodNotAllowed(w http.ResponseWriter, allowedMethod string) {
	w.Header().Set("Allow", allowedMethod)
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func getEnv(key, defaultValue string) string {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	return v
}