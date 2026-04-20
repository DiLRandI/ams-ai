package httpx

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

func StartCSV(w http.ResponseWriter, filename string) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
}

func FormatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(DateLayout)
}

func FormatFloat(v *float64) string {
	if v == nil {
		return ""
	}
	return strconv.FormatFloat(*v, 'f', 2, 64)
}

func FormatInt(v *int) string {
	if v == nil {
		return ""
	}
	return strconv.Itoa(*v)
}
