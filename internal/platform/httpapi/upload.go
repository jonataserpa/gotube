package httpapi

import "net/http"

func createUploadSession(w http.ResponseWriter, r *http.Request, directory Directory) {
	if directory == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "o serviço de upload não está disponível")
		return
	}
	token, ok := bearerToken(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "é preciso entrar para enviar um vídeo")
		return
	}
	var body struct {
		DeclaredSize int64 `json:"declared_size"`
	}
	if err := decodeStrictJSON(r, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "os dados do envio não são válidos")
		return
	}
	session, err := directory.CreateUploadSession(r.Context(), token, r.PathValue("id"), body.DeclaredSize)
	if err != nil {
		writeDirectoryError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, session)
}

func completeUploadSession(w http.ResponseWriter, r *http.Request, directory Directory) {
	if directory == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "o serviço de upload não está disponível")
		return
	}
	token, ok := bearerToken(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "é preciso entrar para confirmar o envio")
		return
	}
	var body struct{}
	if err := decodeStrictJSON(r, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "os dados do envio não são válidos")
		return
	}
	result, err := directory.CompleteUploadSession(r.Context(), token, r.PathValue("id"))
	if err != nil {
		writeDirectoryError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, result)
}
