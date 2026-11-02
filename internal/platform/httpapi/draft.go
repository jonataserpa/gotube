package httpapi

import "net/http"

func createDraft(w http.ResponseWriter, r *http.Request, directory Directory) {
	if directory == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "o serviço de conta não está disponível")
		return
	}
	token, ok := bearerToken(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "é preciso entrar para criar um vídeo")
		return
	}
	var body struct {
		Title string `json:"title"`
	}
	if err := decodeStrictJSON(r, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "os dados do vídeo não são válidos")
		return
	}
	draft, err := directory.CreateDraft(r.Context(), token, body.Title)
	if err != nil {
		writeDirectoryError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, draft)
}
