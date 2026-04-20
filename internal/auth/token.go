package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"strconv"
	"strings"
	"time"

	"ams-ai/internal/domain"
)

func (s *Service) signToken(userID int64, role string, expiresAt time.Time) (string, error) {
	payload := strconv.FormatInt(userID, 10) + "|" + role + "|" + strconv.FormatInt(expiresAt.Unix(), 10)
	mac := hmac.New(sha256.New, []byte(s.secret))
	if _, err := mac.Write([]byte(payload)); err != nil {
		return "", err
	}
	sig := mac.Sum(nil)
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func (s *Service) verifyToken(token string) (int64, string, time.Time, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return 0, "", time.Time{}, domain.ErrUnauthorized
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return 0, "", time.Time{}, domain.ErrUnauthorized
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, "", time.Time{}, domain.ErrUnauthorized
	}
	mac := hmac.New(sha256.New, []byte(s.secret))
	_, _ = mac.Write(payloadBytes)
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return 0, "", time.Time{}, domain.ErrUnauthorized
	}
	fields := strings.Split(string(payloadBytes), "|")
	if len(fields) != 3 {
		return 0, "", time.Time{}, domain.ErrUnauthorized
	}
	userID, err := strconv.ParseInt(fields[0], 10, 64)
	if err != nil {
		return 0, "", time.Time{}, domain.ErrUnauthorized
	}
	role := fields[1]
	if !domain.IsValidRole(role) {
		return 0, "", time.Time{}, domain.ErrUnauthorized
	}
	exp, err := strconv.ParseInt(fields[2], 10, 64)
	if err != nil {
		return 0, "", time.Time{}, domain.ErrUnauthorized
	}
	return userID, role, time.Unix(exp, 0), nil
}

func BearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
