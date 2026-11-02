package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strconv"

	"example.com/gotube/internal/catalog"
)

const (
	defaultVideoLimit = 12
	maxVideoLimit     = 50
)

// PublicVideos lists the public catalog. A nil value means the database is not configured.
type PublicVideos interface {
	ListPublic(ctx context.Context, limit int, after *catalog.Position, channelID string) (catalog.Page, error)
}

var channelIDPattern = regexp.MustCompile(`(?i)^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type apiError struct {
	Error apiErrorBody `json:"error"`
}

type apiErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func listVideos(w http.ResponseWriter, r *http.Request, videos PublicVideos) {
	limit, err := parseLimit(r.URL.Query().Get("limit"))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_LIMIT", "limit deve ser um inteiro entre 1 e 50")
		return
	}
	position, err := parseCursor(r.URL.Query().Get("cursor"))
	if err != nil {
		writeAPIError(w, http.StatusBadRequest, "INVALID_CURSOR", "cursor não é reconhecido")
		return
	}
	channelID := r.URL.Query().Get("channel_id")
	if channelID != "" && !channelIDPattern.MatchString(channelID) {
		writeAPIError(w, http.StatusBadRequest, "INVALID_CHANNEL", "channel_id não é reconhecido")
		return
	}
	if videos == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "o catálogo não está disponível")
		return
	}
	page, err := videos.ListPublic(r.Context(), limit, position, channelID)
	if err != nil {
		writeAPIError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "o catálogo não está disponível")
		return
	}
	writeJSON(w, http.StatusOK, page)
}

func parseLimit(raw string) (int, error) {
	if raw == "" {
		return defaultVideoLimit, nil
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit < 1 || limit > maxVideoLimit {
		return 0, errors.New("limit inválido")
	}
	return limit, nil
}

func parseCursor(raw string) (*catalog.Position, error) {
	if raw == "" {
		return nil, nil
	}
	position, err := catalog.DecodeCursor(raw)
	if err != nil {
		return nil, err
	}
	return &position, nil
}

func writeAPIError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, apiError{Error: apiErrorBody{Code: code, Message: message}})
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "não foi possível codificar a resposta", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}
