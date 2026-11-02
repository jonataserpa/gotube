package httpapi

import "net/http"

func readProcessing(w http.ResponseWriter, r *http.Request, directory Directory) {
	if directory == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "o serviço de processamento não está disponível")
		return
	}
	token, ok := bearerToken(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "é preciso entrar para ver o processamento")
		return
	}
	view, err := directory.ReadProcessing(r.Context(), token, r.PathValue("id"))
	if err != nil {
		writeDirectoryError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, view)
}

func retryProcessing(w http.ResponseWriter, r *http.Request, directory Directory) {
	if directory == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "o serviço de processamento não está disponível")
		return
	}
	token, ok := bearerToken(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "é preciso entrar para tentar de novo")
		return
	}
	var body struct{}
	if err := decodeStrictJSON(r, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "os dados do envio não são válidos")
		return
	}
	result, err := directory.RetryProcessing(r.Context(), token, r.PathValue("id"))
	if err != nil {
		writeDirectoryError(w, err)
		return
	}
	writeJSON(w, http.StatusAccepted, result)
}
