package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"example.com/gotube/internal/identity"
)

func registerAccount(w http.ResponseWriter, r *http.Request, accounts Accounts) {
	if accounts == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "o serviço de conta não está disponível")
		return
	}
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeStrictJSON(r, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "os dados da conta não são válidos")
		return
	}
	account, err := accounts.Register(r.Context(), body.Email, body.Password)
	if err != nil {
		writeAccountError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"id": account.ID, "email": account.Email})
}

func loginAccount(w http.ResponseWriter, r *http.Request, accounts Accounts) {
	if accounts == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "o serviço de conta não está disponível")
		return
	}
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeStrictJSON(r, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "os dados da conta não são válidos")
		return
	}
	session, err := accounts.Login(r.Context(), body.Email, body.Password)
	if err != nil {
		writeAccountError(w, err)
		return
	}
	setRefreshCookie(w, r, session.RefreshToken, session.CookieSecure)
	writeJSON(w, http.StatusOK, sessionJSON(session))
}

func forgotPassword(w http.ResponseWriter, r *http.Request, accounts Accounts) {
	if accounts == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "o serviço de conta não está disponível")
		return
	}
	var body struct {
		Email string `json:"email"`
	}
	if err := decodeStrictJSON(r, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "os dados da conta não são válidos")
		return
	}
	token, err := accounts.Forgot(r.Context(), body.Email)
	if err != nil {
		writeAccountError(w, err)
		return
	}
	payload := map[string]string{
		"message": "Se existir uma conta para esse e-mail, a recuperação foi registrada.",
	}
	if token != "" {
		payload["reset_token"] = token
	}
	writeJSON(w, http.StatusAccepted, payload)
}

func resetPassword(w http.ResponseWriter, r *http.Request, accounts Accounts) {
	if accounts == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "o serviço de conta não está disponível")
		return
	}
	var body struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := decodeStrictJSON(r, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "os dados da conta não são válidos")
		return
	}
	if err := accounts.Reset(r.Context(), body.Token, body.Password); err != nil {
		writeAccountError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func refreshSession(w http.ResponseWriter, r *http.Request, accounts Accounts) {
	if accounts == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "o serviço de conta não está disponível")
		return
	}
	session, err := accounts.Refresh(r.Context(), refreshCookie(r))
	if err != nil {
		writeAccountError(w, err)
		return
	}
	setRefreshCookie(w, r, session.RefreshToken, session.CookieSecure)
	writeJSON(w, http.StatusOK, sessionJSON(session))
}

func logoutSession(w http.ResponseWriter, r *http.Request, accounts Accounts) {
	if accounts == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "o serviço de conta não está disponível")
		return
	}
	if err := accounts.Logout(r.Context(), refreshCookie(r)); err != nil {
		writeAccountError(w, err)
		return
	}
	clearRefreshCookie(w, r, accounts.SecureCookie())
	w.WriteHeader(http.StatusNoContent)
}

func refreshCookie(r *http.Request) string {
	cookie, err := r.Cookie("gotube_refresh")
	if err != nil {
		return ""
	}
	return cookie.Value
}

func clearRefreshCookie(w http.ResponseWriter, r *http.Request, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     "gotube_refresh",
		Value:    "",
		Path:     "/api/auth",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure || r.TLS != nil,
	})
}

func setRefreshCookie(w http.ResponseWriter, r *http.Request, value string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     "gotube_refresh",
		Value:    value,
		Path:     "/api/auth",
		MaxAge:   30 * 24 * 60 * 60,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure || r.TLS != nil,
	})
}

func sessionJSON(session identity.Session) map[string]any {
	return map[string]any{
		"access_token": session.AccessToken,
		"token_type":   "Bearer",
		"expires_in":   session.ExpiresIn,
		"user": map[string]string{
			"id":    session.Account.ID,
			"email": session.Account.Email,
		},
	}
}

func decodeStrictJSON(r *http.Request, dest any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(dest); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("corpo com conteúdo extra")
	}
	return nil
}

func writeAccountError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, identity.ErrValidation):
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "os dados da conta não são válidos")
	case errors.Is(err, identity.ErrEmailTaken):
		writeAPIError(w, http.StatusConflict, "EMAIL_ALREADY_REGISTERED", "este e-mail já está cadastrado")
	case errors.Is(err, identity.ErrInvalidCredentials):
		writeAPIError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "e-mail ou senha não conferem")
	case errors.Is(err, identity.ErrInvalidRefresh):
		writeAPIError(w, http.StatusUnauthorized, "INVALID_REFRESH", "a sessão não é mais válida")
	case errors.Is(err, identity.ErrInvalidReset):
		writeAPIError(w, http.StatusBadRequest, "INVALID_RESET_TOKEN", "o token de recuperação não é válido")
	case errors.Is(err, identity.ErrAuthUnavailable):
		writeAPIError(w, http.StatusServiceUnavailable, "AUTH_UNAVAILABLE", "o serviço de conta não está configurado")
	default:
		writeAPIError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "o serviço de conta não está disponível")
	}
}
