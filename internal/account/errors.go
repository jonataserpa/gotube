package account

import "errors"

var (
	ErrValidation      = errors.New("dados inválidos")
	ErrUnauthenticated = errors.New("não autenticado")
	ErrForbidden       = errors.New("canal de outra conta")
	ErrNotFound        = errors.New("canal não encontrado")
	ErrChannelExists   = errors.New("canal já existe")
	ErrChannelRequired = errors.New("canal obrigatório")
	ErrHandleTaken     = errors.New("handle já cadastrado")
	ErrVideoForbidden  = errors.New("vídeo de outra conta")
	ErrVideoNotFound   = errors.New("vídeo não encontrado")
	ErrNotWaiting      = errors.New("vídeo não aguarda envio")
	ErrUploadRejected  = errors.New("objeto recusado")
	ErrUploadNotFound  = errors.New("sessão não encontrada")
	ErrNotFailed       = errors.New("vídeo não falhou")
	ErrUnavailable     = errors.New("banco indisponível")
)
