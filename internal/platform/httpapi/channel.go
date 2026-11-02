package httpapi

import (
	"net/http"

	"example.com/gotube/internal/account"
)

func readOwnChannel(w http.ResponseWriter, r *http.Request, directory Directory) {
	if directory == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "o serviço de conta não está disponível")
		return
	}
	token, ok := bearerToken(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "é preciso entrar para ver o canal")
		return
	}
	owned, err := directory.ReadOwnChannel(r.Context(), token)
	if err != nil {
		writeDirectoryError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, owned)
}

func createChannel(w http.ResponseWriter, r *http.Request, directory Directory) {
	if directory == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "o serviço de conta não está disponível")
		return
	}
	token, ok := bearerToken(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "é preciso entrar para criar o canal")
		return
	}
	var body channelBody
	if err := decodeStrictJSON(r, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "os dados do canal não são válidos")
		return
	}
	channel, err := directory.CreateChannel(r.Context(), token, body.input())
	if err != nil {
		writeDirectoryError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, channel)
}

func updateChannel(w http.ResponseWriter, r *http.Request, directory Directory) {
	if directory == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "o serviço de conta não está disponível")
		return
	}
	token, ok := bearerToken(r)
	if !ok {
		writeAPIError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "é preciso entrar para editar o canal")
		return
	}
	var body channelPatchBody
	if err := decodeStrictJSON(r, &body); err != nil {
		writeAPIError(w, http.StatusBadRequest, "VALIDATION_ERROR", "os dados do canal não são válidos")
		return
	}
	channel, err := directory.UpdateChannel(r.Context(), token, r.PathValue("id"), body.patch())
	if err != nil {
		writeDirectoryError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, channel)
}

func readPublicChannel(w http.ResponseWriter, r *http.Request, directory Directory) {
	if directory == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "o serviço de conta não está disponível")
		return
	}
	channel, err := directory.ReadPublicChannel(r.Context(), r.PathValue("handle"))
	if err != nil {
		writeDirectoryError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, channel)
}

type channelBody struct {
	Handle      string   `json:"handle"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	CategoryIDs []string `json:"category_ids"`
}

func (b channelBody) input() account.ChannelInput {
	return account.ChannelInput{Handle: b.Handle, Name: b.Name, Description: b.Description, CategoryIDs: b.CategoryIDs}
}

type channelPatchBody struct {
	Handle      *string   `json:"handle"`
	Name        *string   `json:"name"`
	Description *string   `json:"description"`
	CategoryIDs *[]string `json:"category_ids"`
}

func (b channelPatchBody) patch() account.ChannelPatch {
	return account.ChannelPatch{Handle: b.Handle, Name: b.Name, Description: b.Description, CategoryIDs: b.CategoryIDs}
}

func listCategories(w http.ResponseWriter, r *http.Request, directory Directory) {
	if directory == nil {
		writeAPIError(w, http.StatusServiceUnavailable, "DATABASE_UNAVAILABLE", "o serviço de conta não está disponível")
		return
	}
	items, err := directory.ListCategories(r.Context())
	if err != nil {
		writeDirectoryError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}
