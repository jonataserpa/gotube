package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"example.com/gotube/internal/account"
)

// Directory serves profile and channel routes. A nil value means the database is not configured.
type Directory interface {
	ReadProfile(ctx context.Context, rawToken string) (account.Profile, error)
	UpdateProfile(ctx context.Context, rawToken string, displayName string) (account.Profile, error)
	ReadOwnChannel(ctx context.Context, rawToken string) (account.OwnChannel, error)
	CreateChannel(ctx context.Context, rawToken string, input account.ChannelInput) (account.Channel, error)
	UpdateChannel(ctx context.Context, rawToken string, id string, patch account.ChannelPatch) (account.Channel, error)
	ReadPublicChannel(ctx context.Context, handle string) (account.Channel, error)
	ListCategories(ctx context.Context) ([]account.Category, error)
	CreateDraft(ctx context.Context, rawToken string, title string) (account.Draft, error)
	CreateUploadSession(ctx context.Context, rawToken string, videoID string, declaredSize int64) (account.UploadSession, error)
	CompleteUploadSession(ctx context.Context, rawToken string, sessionID string) (account.UploadResult, error)
	ReadProcessing(ctx context.Context, rawToken string, videoID string) (account.ProcessingView, error)
	RetryProcessing(ctx context.Context, rawToken string, videoID string) (account.UploadResult, error)
}

func readProfile(w http.ResponseWriter, r *http.Request, directory Directory) {
	if directory == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "o serviço de conta não está disponível")
		return
	}
	token, ok := bearerToken(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "é preciso entrar para ver o perfil")
		return
	}
	profile, err := directory.ReadProfile(r.Context(), token)
	if err != nil {
		writeDirectoryError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func updateProfile(w http.ResponseWriter, r *http.Request, directory Directory) {
	if directory == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "o serviço de conta não está disponível")
		return
	}
	token, ok := bearerToken(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "é preciso entrar para editar o perfil")
		return
	}
	var body struct {
		DisplayName string `json:"display_name"`
	}
	if err := decodeStrictJSON(r, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "os dados do perfil não são válidos")
		return
	}
	profile, err := directory.UpdateProfile(r.Context(), token, body.DisplayName)
	if err != nil {
		writeDirectoryError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func bearerToken(r *http.Request) (string, bool) {
	value := r.Header.Get("Authorization")
	token, ok := strings.CutPrefix(value, "Bearer ")
	if !ok || strings.TrimSpace(token) == "" || strings.Contains(token, " ") {
		return "", false
	}
	return token, true
}

func writeDirectoryError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, account.ErrValidation):
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "os dados não são válidos")
	case errors.Is(err, account.ErrUnauthenticated):
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "é preciso entrar novamente")
	case errors.Is(err, account.ErrForbidden):
		writeAPIError(w, http.StatusForbidden, "CHANNEL_FORBIDDEN", "esse canal não é desta conta")
	case errors.Is(err, account.ErrNotFound):
		writeAPIError(w, http.StatusNotFound, "CHANNEL_NOT_FOUND", "canal não encontrado")
	case errors.Is(err, account.ErrChannelExists):
		writeAPIError(w, http.StatusConflict, "CHANNEL_ALREADY_EXISTS", "esta conta já tem um canal")
	case errors.Is(err, account.ErrChannelRequired):
		writeAPIError(w, http.StatusConflict, "CHANNEL_REQUIRED", "crie um canal antes de enviar um vídeo")
	case errors.Is(err, account.ErrVideoForbidden):
		writeAPIError(w, http.StatusForbidden, "VIDEO_FORBIDDEN", "esse vídeo não é desta conta")
	case errors.Is(err, account.ErrVideoNotFound):
		writeAPIError(w, http.StatusNotFound, "VIDEO_NOT_FOUND", "vídeo não encontrado")
	case errors.Is(err, account.ErrUploadNotFound):
		writeAPIError(w, http.StatusNotFound, "UPLOAD_SESSION_NOT_FOUND", "sessão de envio não encontrada")
	case errors.Is(err, account.ErrNotWaiting):
		writeAPIError(w, http.StatusConflict, "VIDEO_NOT_WAITING", "esse vídeo não está aguardando envio")
	case errors.Is(err, account.ErrUploadRejected):
		writeAPIError(w, http.StatusConflict, "UPLOAD_REJECTED", "o arquivo enviado não confere com a autorização")
	case errors.Is(err, account.ErrNotFailed):
		writeAPIError(w, http.StatusConflict, "PROCESSING_NOT_FAILED", "esse vídeo não falhou")
	case errors.Is(err, account.ErrHandleTaken):
		writeAPIError(w, http.StatusConflict, "HANDLE_ALREADY_REGISTERED", "esse endereço de canal já está em uso")
	default:
		writeAPIError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "o serviço de conta não está disponível")
	}
}
