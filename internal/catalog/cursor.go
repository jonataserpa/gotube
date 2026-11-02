package catalog

import (
	"encoding/base64"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

// EncodeCursor hides the keyset position. It is not a signed token.
func EncodeCursor(publishedAt time.Time, id string) string {
	payload := publishedAt.UTC().Format(time.RFC3339Nano) + "|" + id
	return base64.RawURLEncoding.EncodeToString([]byte(payload))
}

// DecodeCursor rejects anything that is not a published_at and a UUID.
func DecodeCursor(raw string) (Position, error) {
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return Position{}, errors.New("cursor malformado")
	}
	published, id, ok := strings.Cut(string(decoded), "|")
	if !ok || !validUUID(id) {
		return Position{}, errors.New("cursor malformado")
	}
	publishedAt, err := time.Parse(time.RFC3339Nano, published)
	if err != nil {
		return Position{}, errors.New("cursor malformado")
	}
	return Position{PublishedAt: publishedAt.UTC(), ID: id}, nil
}

func validUUID(value string) bool {
	var parsed pgtype.UUID
	return parsed.Scan(value) == nil && parsed.Valid
}
