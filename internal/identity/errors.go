package identity

import "errors"

var (
	ErrValidation         = errors.New("dados inválidos")
	ErrEmailTaken         = errors.New("e-mail já cadastrado")
	ErrInvalidCredentials = errors.New("credenciais inválidas")
	ErrInvalidRefresh     = errors.New("refresh inválido")
	ErrInvalidReset       = errors.New("reset inválido")
	ErrUnavailable        = errors.New("banco indisponível")
	ErrAuthUnavailable    = errors.New("assinatura indisponível")
)
