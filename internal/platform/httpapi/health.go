package httpapi

import (
	"context"
	"encoding/json"
	"net/http"

	"example.com/gotube/internal/identity"
)

type healthResponse struct {
	Status string `json:"status"`
}

// Accounts is the identity surface. A nil value means the database is not configured.
type Accounts interface {
	Register(ctx context.Context, email string, password string) (identity.Account, error)
	Login(ctx context.Context, email string, password string) (identity.Session, error)
	Refresh(ctx context.Context, refreshToken string) (identity.Session, error)
	Logout(ctx context.Context, refreshToken string) error
	Forgot(ctx context.Context, email string) (string, error)
	Reset(ctx context.Context, token string, password string) error
	SecureCookie() bool
}

// NewMux exposes the routes implemented so far.
// videos or accounts may be nil; the matching routes then report that the database is unavailable.
func NewMux(videos PublicVideos, accounts Accounts, directory Directory) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", Health)
	mux.HandleFunc("GET /v1/videos", func(w http.ResponseWriter, r *http.Request) {
		listVideos(w, r, videos)
	})
	mux.HandleFunc("POST /v1/videos", func(w http.ResponseWriter, r *http.Request) {
		createDraft(w, r, directory)
	})
	mux.HandleFunc("POST /v1/videos/{id}/upload-sessions", func(w http.ResponseWriter, r *http.Request) {
		createUploadSession(w, r, directory)
	})
	mux.HandleFunc("POST /v1/upload-sessions/{id}/complete", func(w http.ResponseWriter, r *http.Request) {
		completeUploadSession(w, r, directory)
	})
	mux.HandleFunc("GET /v1/videos/{id}/processing", func(w http.ResponseWriter, r *http.Request) {
		readProcessing(w, r, directory)
	})
	mux.HandleFunc("POST /v1/videos/{id}/retry-processing", func(w http.ResponseWriter, r *http.Request) {
		retryProcessing(w, r, directory)
	})
	mux.HandleFunc("POST /v1/auth/register", func(w http.ResponseWriter, r *http.Request) {
		registerAccount(w, r, accounts)
	})
	mux.HandleFunc("POST /v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		loginAccount(w, r, accounts)
	})
	mux.HandleFunc("POST /v1/auth/refresh", func(w http.ResponseWriter, r *http.Request) {
		refreshSession(w, r, accounts)
	})
	mux.HandleFunc("POST /v1/auth/logout", func(w http.ResponseWriter, r *http.Request) {
		logoutSession(w, r, accounts)
	})
	mux.HandleFunc("POST /v1/auth/password/forgot", func(w http.ResponseWriter, r *http.Request) {
		forgotPassword(w, r, accounts)
	})
	mux.HandleFunc("POST /v1/auth/password/reset", func(w http.ResponseWriter, r *http.Request) {
		resetPassword(w, r, accounts)
	})
	mux.HandleFunc("GET /v1/me", func(w http.ResponseWriter, r *http.Request) {
		readProfile(w, r, directory)
	})
	mux.HandleFunc("PATCH /v1/me", func(w http.ResponseWriter, r *http.Request) {
		updateProfile(w, r, directory)
	})
	mux.HandleFunc("GET /v1/me/channel", func(w http.ResponseWriter, r *http.Request) {
		readOwnChannel(w, r, directory)
	})
	mux.HandleFunc("POST /v1/channels", func(w http.ResponseWriter, r *http.Request) {
		createChannel(w, r, directory)
	})
	mux.HandleFunc("PATCH /v1/channels/{id}", func(w http.ResponseWriter, r *http.Request) {
		updateChannel(w, r, directory)
	})
	mux.HandleFunc("GET /v1/channels/{handle}", func(w http.ResponseWriter, r *http.Request) {
		readPublicChannel(w, r, directory)
	})
	mux.HandleFunc("GET /v1/categories", func(w http.ResponseWriter, r *http.Request) {
		listCategories(w, r, directory)
	})
	return mux
}

// Health reports process liveness. It does not check PostgreSQL.
func Health(w http.ResponseWriter, _ *http.Request) {
	body, err := json.Marshal(healthResponse{Status: "ok"})
	if err != nil {
		http.Error(w, "não foi possível codificar a resposta", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}
