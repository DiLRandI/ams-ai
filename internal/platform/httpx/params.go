package httpx

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ams-ai/internal/domain"
)

const DateLayout = "2006-01-02"

func PathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		WriteError(w, fmt.Errorf("%w: invalid id", domain.ErrInvalid))
		return 0, false
	}
	return id, true
}

func ParseOptionalDate(raw string) (*time.Time, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	t, err := time.Parse(DateLayout, raw)
	if err != nil {
		return nil, fmt.Errorf("%w: dates must use YYYY-MM-DD", domain.ErrInvalid)
	}
	return &t, nil
}

func ParseRequiredDate(raw, field string) (time.Time, error) {
	t, err := ParseOptionalDate(raw)
	if err != nil {
		return time.Time{}, err
	}
	if t == nil {
		return time.Time{}, fmt.Errorf("%w: %s is required", domain.ErrInvalid, field)
	}
	return *t, nil
}

func NormalizeOptionalStatus(status string) string {
	if strings.TrimSpace(status) == "" {
		return ""
	}
	return domain.NormalizeStatus(status)
}
